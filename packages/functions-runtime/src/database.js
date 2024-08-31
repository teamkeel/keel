const { Kysely, PostgresDialect, CamelCasePlugin } = require("kysely");
const neonserverless = require("@neondatabase/serverless");
const { AsyncLocalStorage } = require("async_hooks");
const { AuditContextPlugin } = require("./auditing");
const pg = require("pg");
const { PROTO_ACTION_TYPES } = require("./consts");
const { withSpan } = require("./tracing");
const ws = require("ws");

// withDatabase is responsible for setting the correct database client in our AsyncLocalStorage
// so that the the code in a custom function uses the correct client.
// For GET and LIST action types, no transaction is used, but for
// actions that mutate data such as CREATE, DELETE & UPDATE, all of the code inside
// the user's custom function is wrapped in a transaction so we can rollback
// the transaction if something goes wrong.
// withDatabase shouldn't be exposed in the public api of the sdk
async function withDatabase(db, actionType, cb) {
  let requiresTransaction = true;

  switch (actionType) {
    case PROTO_ACTION_TYPES.SUBSCRIBER:
    case PROTO_ACTION_TYPES.JOB:
    case PROTO_ACTION_TYPES.GET:
    case PROTO_ACTION_TYPES.LIST:
      requiresTransaction = false;
      break;
  }

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

const dbInstance = new AsyncLocalStorage();

// used to establish a singleton for our vitest environment
let vitestDb = null;

// useDatabase will retrieve the database client set by withDatabase from the local storage
function useDatabase() {
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
  throw new Error("useDatabase must be called within a function");
}

// createDatabaseClient will return a brand new instance of Kysely. Every instance of Kysely
// represents an individual connection to the database.
// not to be exported externally from our sdk - consumers should use useDatabase
function createDatabaseClient() {
  const db = new Kysely({
    dialect: getDialect(),
    plugins: [
      // ensures that the audit context data is written to Postgres configuration parameters
      new AuditContextPlugin(),
      // allows users to query using camelCased versions of the database column names, which
      // should match the names we use in our schema.
      // https://kysely-org.github.io/kysely/classes/CamelCasePlugin.html
      // If they don't, then we can create a custom implementation of the plugin where we control
      // the casing behaviour (see url above for example)
      new CamelCasePlugin(),
    ],
    log(event) {
      if ("DEBUG" in process.env) {
        if (event.level === "query") {
          console.log(event.query.sql);
          console.log(event.query.parameters);
        }
      }
    },
  });

  return db;
}

class InstrumentedPool extends pg.Pool {
  async connect(...args) {
    const _super = super.connect.bind(this);
    return withSpan("Database Connect", function (span) {
      span.setAttribute("dialect", process.env["KEEL_DB_CONN_TYPE"]);
      return _super(...args);
    });
  }
}

class InstrumentedNeonServerlessPool extends neonserverless.Pool {
  async connect(...args) {
    const _super = super.connect.bind(this);
    return withSpan("Database Connect", function (span) {
      span.setAttribute("dialect", process.env["KEEL_DB_CONN_TYPE"]);
      return _super(...args);
    });
  }
}

const txStatements = {
  begin: "Transaction Begin",
  commit: "Transaction Commit",
  rollback: "Transaction Rollback",
};

class InstrumentedClient extends pg.Client {
  async query(...args) {
    const _super = super.query.bind(this);
    const sql = args[0];

    let sqlAttribute = false;
    let spanName = txStatements[sql.toLowerCase()];
    if (!spanName) {
      spanName = "Database Query";
      sqlAttribute = true;
    }

    return withSpan(spanName, function (span) {
      if (sqlAttribute) {
        span.setAttribute("sql", args[0]);
        span.setAttribute("dialect", process.env["KEEL_DB_CONN_TYPE"]);
      }
      return _super(...args);
    });
  }
}

function getDialect() {
  // Adding a custom type parser for numeric fields: see https://kysely.dev/docs/recipes/data-types#configuring-runtime-javascript-types
  // 1700 = type for NUMERIC
  pg.types.setTypeParser(1700, function (val) {
    return parseFloat(val);
  });

  const dbConnType = process.env["KEEL_DB_CONN_TYPE"];
  switch (dbConnType) {
    case "pg":
      return new PostgresDialect({
        pool: new InstrumentedPool({
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
          connectionString: mustEnv("KEEL_DB_CONN"),
        }),
      });
    case "neon":
      neonserverless.neonConfig.webSocketConstructor = ws;

      const pool = new InstrumentedNeonServerlessPool({
        connectionString: mustEnv("KEEL_DB_CONN"),
      });

      pool.on("connect", (client) => {
        const originalQuery = client.query;
        client.query = function (...args) {
          const sql = args[0];

          let sqlAttribute = false;
          let spanName = txStatements[sql.toLowerCase()];
          if (!spanName) {
            spanName = "Database Query";
            sqlAttribute = true;
          }

          return withSpan(spanName, function (span) {
            if (sqlAttribute) {
              span.setAttribute("sql", args[0]);
              span.setAttribute("dialect", process.env["KEEL_DB_CONN_TYPE"]);
            }
            return originalQuery.apply(client, args);
          });
        };
      });

      return new PostgresDialect({
        pool: pool,
      });
    default:
      throw Error("unexpected KEEL_DB_CONN_TYPE: " + dbConnType);
  }
}

function mustEnv(key) {
  const v = process.env[key];
  if (!v) {
    throw new Error(`expected environment variable ${key} to be set`);
  }
  return v;
}

module.exports = {
  createDatabaseClient,
  useDatabase,
  withDatabase,
};
