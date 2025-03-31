const { withDatabase } = require("./database");

function tryExecuteFlow({ db, permitted, request, functionConfig }, cb) {
  return withDatabase(db, false, async () => {
    return cb();
  });
}

module.exports.tryExecuteFlow = tryExecuteFlow;
