const { withDatabase } = require("./database");

function tryExecuteFlow(db, cb) {
  return withDatabase(db, false, async () => {
    return cb();
  });
}

module.exports.tryExecuteFlow = tryExecuteFlow;
