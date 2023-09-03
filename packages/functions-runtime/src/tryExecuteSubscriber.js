const { withDatabase } = require("./database");
const { withAuditContext } = require("./auditing");

// tryExecuteSubscriber will create a new database connection and execute the function call.
function tryExecuteSubscriber({ db, actionType }, cb) {
  return withDatabase(db, actionType, async () => {
      withAuditContext(request, async () => {
        return await cb();
      });
  });
}

module.exports.tryExecuteSubscriber = tryExecuteSubscriber;
