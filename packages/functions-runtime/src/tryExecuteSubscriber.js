const { withDatabase } = require("./database");

// tryExecuteSubscriber will create a new database connection and execute the function call.
function tryExecuteSubscriber({ db, actionType }, cb) {
  return withDatabase(db, actionType, async () => {
    await cb();
  });
}

module.exports.tryExecuteSubscriber = tryExecuteSubscriber;
