import { withDatabase } from "./database";
import { withAuditContext } from "./auditing";
import { withPermissions, PERMISSION_STATE } from "./permissions";
import { PermissionError } from "./errors";

// tryExecuteJob will create a new database transaction around a function call
// and handle any permissions checks. If a permission check fails, then an Error will be thrown and the catch block will be hit.
function tryExecuteJob({ db, permitted, request, functionConfig }, cb) {
  return withPermissions(permitted, async ({ getPermissionState }) => {
    let requiresTransaction = false;
    if (functionConfig?.dbTransaction !== undefined) {
      requiresTransaction = functionConfig.dbTransaction;
    }
    return withDatabase(db, requiresTransaction, async () => {
      await withAuditContext(request, async () => {
        return cb();
      });

      // api.permissions maintains an internal state of whether the current operation has been *explicitly* permitted/denied by the user in the course of their custom function, or if execution has already been permitted by a role based permission (evaluated in the main runtime).
      // we need to check that the final state is permitted or unpermitted. if it's not, then it means that the user has taken no explicit action to permit/deny
      // and therefore we default to checking the permissions defined in the schema automatically.
      if (getPermissionState() === PERMISSION_STATE.UNPERMITTED) {
        throw new PermissionError(`Not permitted to access ${request.method}`);
      }
    });
  });
}

export { tryExecuteJob };
