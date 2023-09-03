const { withDatabase } = require("./database");
const { withAuditContext } = require("./auditing");
const {
  withPermissions,
  PERMISSION_STATE,
  checkBuiltInPermissions,
} = require("./permissions");
const { PermissionError } = require("./errors");
const { PROTO_ACTION_TYPES } = require("./consts");

// tryExecuteFunction will create a new database transaction around a function call
// and handle any permissions checks. If a permission check fails, then an Error will be thrown and the catch block will be hit.
function tryExecuteFunction(
  { request, db, permitted, permissionFns, actionType, ctx },
  cb
) {
  return withPermissions(permitted, async ({ getPermissionState }) => {
    return withDatabase(db, actionType, async ({ transaction }) => {
      const fnResult = await withAuditContext(request, async () => {
        return cb();
      });

      // api.permissions maintains an internal state of whether the current function has been *explicitly* permitted/denied by the user in the course of their custom function, or if execution has already been permitted by a role based permission (evaluated in the main runtime).
      // we need to check that the final state is permitted or unpermitted. if it's not, then it means that the user has taken no explicit action to permit/deny
      // and therefore we default to checking the permissions defined in the schema automatically.
      switch (getPermissionState()) {
        case PERMISSION_STATE.PERMITTED:
          return fnResult;
        case PERMISSION_STATE.UNPERMITTED:
          throw new PermissionError(
            `Not permitted to access ${request.method}`
          );
        default:
          // unknown state, proceed with checking against the built in permissions in the schema
          const relevantPermissions = permissionFns[request.method];

          const peakInsideTransaction =
            actionType === PROTO_ACTION_TYPES.CREATE;

          let rowsForPermissions = [];
          if (fnResult != null) {
            switch (actionType) {
              case PROTO_ACTION_TYPES.LIST:
                rowsForPermissions = fnResult;
                break;
              case PROTO_ACTION_TYPES.DELETE:
                rowsForPermissions = [{ id: fnResult }];
                break;
              case (PROTO_ACTION_TYPES.GET, PROTO_ACTION_TYPES.CREATE):
                rowsForPermissions = [fnResult];
                break;
              default:
                rowsForPermissions = [fnResult];
                break;
            }
          }

          // check will throw a PermissionError if a permission rule is invalid
          await checkBuiltInPermissions({
            rows: rowsForPermissions,
            permissionFns: relevantPermissions,
            // it is important that we pass db here as db represents the connection to the database
            // *outside* of the current transaction. Given that any changes inside of a transaction
            // are opaque to the outside, we can utilize this when running permission rules and then deciding to
            // rollback any changes if they do not pass. However, for creates we need to be able to 'peak' inside the transaction to read the created record, as this won't exist outside of the transaction.
            db: peakInsideTransaction ? transaction : db,
            ctx,
            functionName: request.method,
          });

          // If the built in permission check above doesn't throw, then it means that the request is permitted and we can continue returning the return value from the custom function out of the transaction
          return fnResult;
      }
    });
  });
}

module.exports.tryExecuteFunction = tryExecuteFunction;
