const { KNOWN_CLIENTS } = require("./clients");
const { AsyncLocalStorage } = require("async_hooks");
const { PROTO_ACTION_TYPES } = require("./consts");
const { getDatabase: getKysely } = require("./clients/kysely");

// withDatabase sets up a new database client that custom functions will utilize.
// The database client is stored in an AsyncLocalStorage store so consumers further down the hierarchy can access the db.
// Create / Update action types require the custom function to be executed inside of a transaction so therefore
// the db client returned will execute any queries inside a transaction 
// For read type operations such as list & get, no transaction is used
async function withDatabase({ actionType, orm = KNOWN_CLIENTS.KYSELY }, cb) {
  const db = await getDatabase(orm);

  let requiresTransaction = true;

  switch (actionType) {
    case PROTO_ACTION_TYPES.GET:
    case PROTO_ACTION_TYPES.LIST:
      requiresTransaction = false;
  }

  if (requiresTransaction) {
    switch (orm) {
      case KNOWN_CLIENTS.KYSELY:
        return db.transaction().execute(async (transaction) => {
          return dbInstance.run(transaction, async () => {
            return cb({ transaction, db });
          });
        });
      case KNOWN_CLIENTS.PRISMA:
    }
  }
  return dbInstance.run(db, async () => {
    return cb({ db, transaction: null });
  });
}

let db = null;
const dbInstance = new AsyncLocalStorage();

// getDatabase will first check for an instance of Kysely in AsyncLocalStorage,
// otherwise it will create a new instance and reuse it..
async function getDatabase(orm = KNOWN_CLIENTS.KYSELY) {
  let fromStore = dbInstance.getStore();
  if (fromStore) {
    return fromStore;
  }

  if (db) {
    return db;
  }

  switch (orm) {
    case KNOWN_CLIENTS.KYSELY:
      db = getKysely();
      break;
    case KNOWN_CLIENTS.PRISMA:
      const { PrismaClient } = await import("@prisma/client");

      db = new PrismaClient()
      break;
  }

  return db;
}

module.exports.getDatabase = getDatabase;
module.exports.withDatabase = withDatabase;
