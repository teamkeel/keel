import { withDatabase } from "./database";
import { withAuditContext } from "./auditing";

// tryExecuteSubscriber will create a new database connection and execute the function call.
function tryExecuteSubscriber({ request, db, functionConfig }, cb) {
  let requiresTransaction = false;
  if (functionConfig?.dbTransaction !== undefined) {
    requiresTransaction = functionConfig.dbTransaction;
  }
  return withDatabase(db, requiresTransaction, async () => {
    await withAuditContext(request, async () => {
      return cb();
    });
  });
}

export { tryExecuteSubscriber };
