import { withDatabase } from "./database";
import { withAuditContext } from "./auditing";

function tryExecuteFlow(db, request, cb) {
  return withDatabase(db, false, async () => {
    return withAuditContext(request, async () => {
      return cb();
    });
  });
}

export { tryExecuteFlow };
