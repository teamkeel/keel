const { withDatabase } = require("./database");
const { withAuditContext } = require("./auditing");

// tryExecuteSubscriber will create a new database connection and execute the function call.
function tryExecuteSubscriber({ db, actionType }, cb) {
  return withDatabase(db, actionType, async () => {
      await withAuditContext(request, async () => {
        return cb();
      });
  });
}

module.exports.tryExecuteSubscriber = tryExecuteSubscriber;
