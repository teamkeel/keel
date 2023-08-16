const { AsyncLocalStorage } = require("async_hooks");
const TraceParent = require("traceparent");

const auditContextStorage = new AsyncLocalStorage();

// withAuditContext creates the audit context from the runtime request body
// and sets it to in AsyncLocalStorage so that this data is available to the
// ModelAPI during the execution of actions, jobs and subscribers.
async function withAuditContext(request, cb) {
  let audit = {};
  if (request.meta?.identity) {
    audit.identityId = request.meta.identity.id;
  }
  if (request.meta?.tracing?.traceparent) {
    audit.traceId = TraceParent.fromString(
      request.meta.tracing.traceparent
    )?.traceId;
  }

  return await auditContextStorage.run(audit, () => {
    return cb();
  });
}

// getAuditContext retrieves the audit context from AsyncLocalStorage.
function getAuditContext() {
  let auditStore = auditContextStorage.getStore();
  return {
    identityId: auditStore?.identityId,
    traceId: auditStore?.traceId,
  };
}

module.exports.withAuditContext = withAuditContext;
module.exports.getAuditContext = getAuditContext;
