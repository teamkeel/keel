import { Kysely, PostgresDialect, KyselyConfig } from "kysely";
import * as neon from "@neondatabase/serverless";
import { AsyncLocalStorage } from "node:async_hooks";
import { AuditContextPlugin } from "./auditing";
import { KeelCamelCasePlugin } from "./camelCasePlugin";
import { Pool, Client, types as pgTypes, PoolConfig, PoolClient } from "pg";
import { withSpan, KEEL_INTERNAL_ATTR } from "./tracing";
import WebSocket from "ws";
import { readFileSync } from "node:fs";
import { Duration } from "./Duration";

interface DatabaseContext {
  transaction?: Kysely<any>;
  sDb?: Kysely<any>;
}

interface DatabaseClientConfig {
  connString?: string;
}

const dbInstance = new AsyncLocalStorage<Kysely<any>>();

// used to establish a singleton for our vitest environment
let vitestDb: Kysely<any> | null = null;

// withDatabase is responsible for setting the correct database client in our AsyncLocalStorage
// so that the the code in a custom function uses the correct client.
// For GET and LIST action types, no transaction is used, but for
// actions that mutate data such as CREATE, DELETE & UPDATE, all of the code inside
// the user's custom function is wrapped in a transaction so we can rollback
// the transaction if something goes wrong.
// withDatabase shouldn't be exposed in the public api of the sdk
async function withDatabase<T>(
  db: Kysely<any>,
  requiresTransaction: boolean,
  cb: (context: DatabaseContext) => Promise<T>
): Promise<T> {
  // db.transaction() provides a kysely instance bound to a transaction.
  if (requiresTransaction) {
    return db.transaction().execute(async (transaction) => {
      return dbInstance.run(transaction, async () => {
        return cb({ transaction });
      });
    });
  }

  // db.connection() provides a kysely instance bound to a single database connection.
  return db.connection().execute(async (sDb) => {
    return dbInstance.run(sDb, async () => {
      return cb({ sDb });
    });
  });
}

// useDatabase will retrieve the database client set by withDatabase from the local storage
function useDatabase(): Kysely<any> {
  // retrieve the instance of the database client from the store which is aware of
  // which context the current connection to the db is running in - e.g does the context
  // require a transaction or not?
  let fromStore = dbInstance.getStore();
  if (fromStore) {
    return fromStore;
  }

  // if the NODE_ENV is 'test' then we know we are inside of the vitest environment
  // which covers any test files ending in *.test.ts. Custom function code runs in a different node process which will not have this environment variable. Tests written using our testing
  // framework call actions (and in turn custom function code) over http using the ActionExecutor class
  if ("NODE_ENV" in process.env && process.env.NODE_ENV == "test") {
    if (!vitestDb) {
      vitestDb = createDatabaseClient();
    }
    return vitestDb;
  }

  // If we've gotten to this point, then we know that we are in a custom function runtime server
  // context and we haven't been able to retrieve the in-context instance of Kysely, which means we should throw an error.
  console.trace();
  throw new Error("useDatabase must be called within a function");
}

// createDatabaseClient will return a brand new instance of Kysely. Every instance of Kysely
// represents an individual connection to the database.
// not to be exported externally from our sdk - consumers should use useDatabase
function createDatabaseClient(config: DatabaseClientConfig = {}): Kysely<any> {
  const kyseleyConfig: KyselyConfig = {
    dialect: getDialect(config.connString),
    plugins: [
      // ensures that the audit context data is written to Postgres configuration parameters
      new AuditContextPlugin(),
      // allows users to query using camelCased versions of the database column names, which
      // should match the names we use in our schema.
      // We're using an extended version of Kysely's CamelCasePlugin which avoids changing keys of objects that represent
      // rich data formats, specific to Keel (e.g. Duration)
      new KeelCamelCasePlugin(),
    ],
    log(event: any) {
      if (process.env.DEBUG) {
        if (event.level === "query") {
          console.log(event.query.sql);
          console.log(event.query.parameters);
        }
      }
    },
  };

  return new Kysely(kyseleyConfig);
}

class InstrumentedPool extends Pool {
  connect(...args: any): Promise<PoolClient> {
    const _super = super.connect.bind(this);
    return withSpan("Database Connect", function (span: any) {
      span.setAttribute("dialect", process.env["KEEL_DB_CONN_TYPE"]);
      span.setAttribute(KEEL_INTERNAL_ATTR, true);
      return _super.apply(null, args);
    });
  }
}

class InstrumentedNeonServerlessPool extends neon.Pool {
  async connect(...args: any): Promise<neon.PoolClient> {
    const _super = super.connect.bind(this);
    return withSpan("Database Connect", function (span: any) {
      span.setAttribute("dialect", process.env["KEEL_DB_CONN_TYPE"]);
      span.setAttribute(KEEL_INTERNAL_ATTR, true);
      return _super.apply(null, args);
    });
  }
}

const txStatements = {
  begin: "Transaction Begin",
  commit: "Transaction Commit",
  rollback: "Transaction Rollback",
};

class InstrumentedClient extends Client {
  async query(...args: any): Promise<any> {
    const _super = super.query.bind(this);
    const sql = args[0];

    let sqlAttribute = false;
    let spanName = txStatements[sql.toLowerCase() as keyof typeof txStatements];
    if (!spanName) {
      spanName = "Database Query";
      sqlAttribute = true;
    }

    return withSpan(spanName, function (span: any) {
      if (sqlAttribute) {
        span.setAttribute("sql", args[0]);
        span.setAttribute("dialect", process.env["KEEL_DB_CONN_TYPE"]);
      }
      return _super.apply(null, args);
    });
  }
}

function getDialect(connString?: string): PostgresDialect {
  const dbConnType = process.env.KEEL_DB_CONN_TYPE;
  switch (dbConnType) {
    case "pg": {
      // Adding a custom type parser for numeric fields: see https://kysely.dev/docs/recipes/data-types#configuring-runtime-javascript-types
      // 1700 = type for NUMERIC
      pgTypes.setTypeParser(pgTypes.builtins.NUMERIC, (val: string) =>
        parseFloat(val)
      );
      // Adding a custom type parser for interval fields: see https://kysely.dev/docs/recipes/data-types#configuring-runtime-javascript-types
      // 1186 = type for INTERVAL
      pgTypes.setTypeParser(
        pgTypes.builtins.INTERVAL,
        (val: string) => new Duration(val)
      );

      const poolConfig: PoolConfig = {
        Client: InstrumentedClient,
        // Increased idle time before closing a connection in the local pool (from 10s default).
        // Establising a new connection on (almost) every functions query can be expensive, so this
        // will reduce having to open connections as regularly. https://node-postgres.com/apis/pool
        //
        // NOTE: We should consider setting this to 0 (i.e. never pool locally) and open and close
        // connections with each invocation. This is because the freeze/thaw nature of lambdas can cause problems
        // with long-lived connections - see https://github.com/brianc/node-postgres/issues/2718
        // Once we're "fully regional" this should not be a performance problem anymore.
        //
        // Although I doubt we will run into these freeze/thaw issues if idleTimeoutMillis is always shorter than the
        // time is takes for a lambda to freeze (which is not a constant, but could be as short as several minutes,
        // https://www.pluralsight.com/resources/blog/cloud/how-long-does-aws-lambda-keep-your-idle-functions-around-before-a-cold-start)
        idleTimeoutMillis: 50000,
        // If connString is not passed fall back to reading from env var
        connectionString: connString || process.env.KEEL_DB_CONN,
      };

      // Allow the setting of a cert (.pem) file. RDS requires this to enforce SSL.
      if (process.env.KEEL_DB_CERT) {
        poolConfig.ssl = { ca: readFileSync(process.env.KEEL_DB_CERT) };
      }

      return new PostgresDialect({
        pool: new InstrumentedPool(poolConfig),
      });
    }
    case "neon": {
      // Adding a custom type parser for numeric fields: see https://kysely.dev/docs/recipes/data-types#configuring-runtime-javascript-types
      // 1700 = type for NUMERIC
      neon.types.setTypeParser(pgTypes.builtins.NUMERIC, (val: string) =>
        parseFloat(val)
      );

      // Adding a custom type parser for interval fields: see https://kysely.dev/docs/recipes/data-types#configuring-runtime-javascript-types
      // 1186 = type for INTERVAL
      neon.types.setTypeParser(
        pgTypes.builtins.INTERVAL,
        (val: string) => new Duration(val)
      );

      neon.neonConfig.webSocketConstructor = WebSocket;

      const pool = new InstrumentedNeonServerlessPool({
        // If connString is not passed fall back to reading from env var
        connectionString: connString || process.env.KEEL_DB_CONN,
      });

      pool.on("connect", (client: any) => {
        const originalQuery = client.query;
        client.query = function (...args: any[]) {
          const sql = args[0];

          let sqlAttribute = false;
          let spanName =
            txStatements[sql.toLowerCase() as keyof typeof txStatements];
          if (!spanName) {
            spanName = "Database Query";
            sqlAttribute = true;
          }

          return withSpan(spanName, function (span: any) {
            if (sqlAttribute) {
              span.setAttribute("sql", args[0]);
              span.setAttribute("dialect", dbConnType);
            }
            return originalQuery.apply(client, args);
          });
        };
      });

      return new PostgresDialect({
        pool: pool,
      });
    }
    default:
      throw Error("unexpected KEEL_DB_CONN_TYPE: " + dbConnType);
  }
}

export {
  createDatabaseClient,
  useDatabase,
  withDatabase,
  type DatabaseContext,
  type DatabaseClientConfig,
};
