import { withDatabase } from "./database";

function tryExecuteFlow(db, cb) {
  return withDatabase(db, false, async () => {
    return cb();
  });
}

export { tryExecuteFlow };
