const { withDatabase } = require("./database");
const { withPermissions, PERMISSION_STATE } = require("./permissions");

const { PermissionError } = require("./errors");

// tryExecuteJob will create a new database transaction around a function call
// and handle any permissions checks. If a permission check fails, then an Error will be thrown and the catch block will be hit.
function tryExecuteJob({ db, permitted, actionType, request }, cb) {
  return withPermissions(permitted, async ({ getPermissionState }) => {
    return withDatabase(db, actionType, async ({ transaction }) => {
      await cb();

      // we need to check that the final state is unpermitted. if it's not, then it means that the user has taken no explicit action to permit/deny
      if (getPermissionState() === PERMISSION_STATE.UNPERMITTED) {
        throw new PermissionError(`Not permitted to access ${request.method}`);
      }
    });
  });
}

module.exports.tryExecuteJob = tryExecuteJob;
