const { Kysely, PostgresDialect } = require("kysely");
const { AsyncLocalStorage } = require("async_hooks");
const pg = require("pg");
const { PROTO_ACTION_TYPES } = require("./consts");

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

  // todo: ideally we wouldn't want to give you a fresh Kysely instance here if nothing
  // has been found in the context, but the @teamkeel/testing package needs some restructuring
  // to allow for the database client to be set in the store so that this method can throw an error at this line instead of returning a fresh kysely instance.
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

module.exports.useDatabase = useDatabase;
module.exports.withDatabase = withDatabase;
