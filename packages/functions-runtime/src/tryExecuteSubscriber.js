const { withDatabase } = require("./database");
const { withAuditContext } = require("./auditing");

// tryExecuteSubscriber will create a new database connection and execute the function call.
function tryExecuteSubscriber({ request, db, functionConfig }, cb) {
  let requiresTransaction = true;
  if (functionConfig?.dbTransaction !== undefined) {
    requiresTransaction = functionConfig.dbTransaction;
  }
  return withDatabase(db, requiresTransaction, async () => {
    await withAuditContext(request, async () => {
      return cb();
    });
  });
}

module.exports.tryExecuteSubscriber = tryExecuteSubscriber;
