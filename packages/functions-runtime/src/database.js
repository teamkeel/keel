const { Kysely, PostgresDialect, CamelCasePlugin } = require("kysely");
const { AsyncLocalStorage } = require("async_hooks");
const pg = require("pg");
const { PROTO_ACTION_TYPES } = require("./consts");
const { withSpan } = require("./tracing");

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
    case PROTO_ACTION_TYPES.GET:
    case PROTO_ACTION_TYPES.LIST:
      requiresTransaction = false;
      break;
  }

  if (requiresTransaction) {
    return db.transaction().execute(async (transaction) => {
      return dbInstance.run(transaction, async () => {
        return cb({ transaction });
      });
    });
  }

  return dbInstance.run(db, async () => {
    return cb({ transaction: db });
  });
}

let db = null;
const dbInstance = new AsyncLocalStorage();

// useDatabase will retrieve the database client set by withDatabase from the local storage
function useDatabase() {
  let fromStore = dbInstance.getStore();
  if (fromStore) {
    return fromStore;
  }

  if (db) {
    return db;
  }

  // if the NODE_ENV is 'test' then we know we are inside of the vitest environment
  // which covers any test files ending in *.test.ts. Custom function code runs in a different node process which will not have this environment variable. Tests written using our testing
  // framework call actions (and in turn custom function code) over http using the ActionExecutor class
  if ("NODE_ENV" in process.env && process.env.NODE_ENV == "test") {
    // Memoize it for next use.
    db = getDatabaseClient();

    return db;
  }

  // If we've gotten to this point, then we know that we are in a custom function runtime server
  // context and we haven't been able to retrieve the in-context instance of Kysely, which means we should throw an error.
  throw new Error("no database client in context");
}

// getDatabaseClient will return a brand new instance of Kysely.
// not to be exported externally from our sdk - consumers should use useDatabase
function getDatabaseClient() {
  return new Kysely({
    dialect: getDialect(),
    plugins: [
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
}

function mustEnv(key) {
  const v = process.env[key];
  if (!v) {
    throw new Error(`expected environment variable ${key} to be set`);
  }
  return v;
}

class InstrumentedPool extends pg.Pool {
  async connect(...args) {
    const _super = super.connect.bind(this);
    return withSpan("Database Connect", function () {
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
      }
      return _super(...args);
    });
  }
}

function getDialect() {
  const dbConnType = process.env["KEEL_DB_CONN_TYPE"];
  switch (dbConnType) {
    case "pg":
      return new PostgresDialect({
        pool: new InstrumentedPool({
          Client: InstrumentedClient,
          connectionString: mustEnv("KEEL_DB_CONN"),
        }),
      });

    default:
      throw Error("unexpected KEEL_DB_CONN_TYPE: " + dbConnType);
  }
}

module.exports.getDatabaseClient = getDatabaseClient;
module.exports.useDatabase = useDatabase;
module.exports.withDatabase = withDatabase;
