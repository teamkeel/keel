const { Kysely, PostgresDialect } = require("kysely");
const { AsyncLocalStorage } = require("async_hooks");
const pg = require("pg");

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
  });

  return db;
}

module.exports.dbInstance = dbInstance;
module.exports.getDatabase = getDatabase;
