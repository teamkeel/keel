
const { AsyncLocalStorage } = require("async_hooks");
const TraceParent = require("traceparent");

const auditContextStorage = new AsyncLocalStorage();

async function withAuditContext({ identityId, traceId }, cb) {
  return await auditContextStorage.run(
    { identityId: identityId, traceId: traceId },
    () => {
      return cb();
    }
  );
}

function getAuditContext() {
    let auditStore = auditContextStorage.getStore();
    return {
      identityId: auditStore?.identityId,
      traceId: auditStore?.traceId  
    };
}

function auditFromRequest(request) {
    let audit = {};
    if (request.meta?.identity) {
      audit.identityId = request.meta.identity.id
    }
    if (request.meta.tracing?.traceparent) {
      audit.traceId = TraceParent.fromString(
        request.meta.tracing.traceparent
      )?.traceId;
    }
    return audit;
}

module.exports.withAuditContext = withAuditContext;
module.exports.getAuditContext = getAuditContext;
module.exports.auditFromRequest = auditFromRequest;
