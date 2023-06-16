const { Kysely, PostgresDialect } = require("kysely");
const { AsyncLocalStorage } = require("async_hooks");
const pg = require("pg");
const { PROTO_ACTION_TYPES } = require("./consts");

// withTransaction wraps the containing code with a transaction
// and sets the transaction in the AsyncLocalStorage so consumers further
// down the hierarchy can access the current transaction.
// For read type operations such as list & get, no transaction is used
async function withTransaction(db, actionType, cb) {
  switch (actionType) {
    case PROTO_ACTION_TYPES.GET:
    case PROTO_ACTION_TYPES.LIST:
      return dbInstance.run(db, async () => {
        return cb({ transaction: db });
      });
    default:
      return db.transaction().execute(async (transaction) => {
        return dbInstance.run(transaction, async () => {
          return cb({ transaction });
        });
      });
  }
}

function mustEnv(key) {
  const v = process.env[key];
  if (!v) {
    throw new Error(`expected environment variable ${key} to be set`);
  }
  return v;
}

function getDialect() {
  const dbConnType = process.env["KEEL_DB_CONN_TYPE"];
  switch (dbConnType) {
    case "pg":
      return new PostgresDialect({
        pool: new pg.Pool({
          connectionString: mustEnv("KEEL_DB_CONN"),
        }),
      });

    default:
      throw Error("unexpected KEEL_DB_CONN_TYPE: " + dbConnType);
  }
}

let db = null;
const dbInstance = new AsyncLocalStorage();

// getDatabase will first check for an instance of Kysely in AsyncLocalStorage,
// otherwise it will create a new instance and reuse it..
function getDatabase() {
  let fromStore = dbInstance.getStore();
  if (fromStore) {
    return fromStore;
  }

  if (db) {
    return db;
  }

  db = new Kysely({
    dialect: getDialect(),
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

module.exports.getDatabase = getDatabase;
module.exports.withTransaction = withTransaction;
