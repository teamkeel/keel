const { KNOWN_CLIENTS } = require('./clients')
const { AsyncLocalStorage } = require("async_hooks");
const { PROTO_ACTION_TYPES } = require("./consts");
const { getDatabase: getKysely } =  require("./clients/kysely")

// withTransaction wraps the containing code with a transaction
// and sets the transaction in the AsyncLocalStorage so consumers further
// down the hierarchy can access the current transaction.
// For read type operations such as list & get, no transaction is used
async function withTransaction({ actionType, orm = KNOWN_CLIENTS.KYSELY }, cb) {
  const db = getDatabase(orm)

  let requiresTransaction = true;

  switch (actionType) {
    case PROTO_ACTION_TYPES.GET:
    case PROTO_ACTION_TYPES.LIST:
      requiresTransaction = false;
  }

  if (requiresTransaction) {
    switch(orm) {
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
function getDatabase(orm = KNOWN_CLIENTS.KYSELY) {
  let fromStore = dbInstance.getStore();
  if (fromStore) {
    return fromStore;
  }

  if (db) {
    return db;
  }


  switch(orm) {
    case KNOWN_CLIENTS.KYSELY:
      db = getKysely()
    case KNOWN_CLIENTS.PRISMA:
      // todo
    default:

  }

  return db
}

module.exports.getDatabase = getDatabase;
module.exports.withTransaction = withTransaction;
