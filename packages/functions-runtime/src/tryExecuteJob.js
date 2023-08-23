const { withDatabase } = require("./database");
const {
  withPermissions,
  PERMISSION_STATE,
  PermissionError,
} = require("./permissions");

// tryExecuteJob will create a new database transaction around a function call
// and handle any permissions checks. If a permission check fails, then an Error will be thrown and the catch block will be hit.
function tryExecuteJob({ db, permitted, actionType, request }, cb) {
  return withPermissions(permitted, async ({ getPermissionState }) => {
    return withDatabase(db, actionType, async ({ transaction }) => {
      await cb();
      // api.permissions maintains an internal state of whether the current operation has been *explicitly* permitted/denied by the user in the course of their custom function, or if execution has already been permitted by a role based permission (evaluated in the main runtime).
      // we need to check that the final state is permitted or unpermitted. if it's not, then it means that the user has taken no explicit action to permit/deny
      // and therefore we default to checking the permissions defined in the schema automatically.
      if (getPermissionState() === PERMISSION_STATE.UNPERMITTED) {
        throw new PermissionError(`Not permitted to access ${request.method}`);
      }
    });
  });
}

module.exports.tryExecuteJob = tryExecuteJob;
