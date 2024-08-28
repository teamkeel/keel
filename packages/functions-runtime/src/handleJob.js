const {
  createJSONRPCErrorResponse,
  createJSONRPCSuccessResponse,
  JSONRPCErrorCode,
} = require("json-rpc-2.0");
const { getDatabaseClient } = require("./database");
const { errorToJSONRPCResponse, RuntimeErrors } = require("./errors");
const opentelemetry = require("@opentelemetry/api");
const { withSpan } = require("./tracing");
const { PROTO_ACTION_TYPES } = require("./consts");
const { tryExecuteJob } = require("./tryExecuteJob");
const { parseParams } = require("./parsing");

// Generic handler function that is agnostic to runtime environment (local or lambda)
// to execute a job function based on the contents of a jsonrpc-2.0 payload object.
// To read more about jsonrpc request and response shapes, please read https://www.jsonrpc.org/specification
async function handleJob(request, config) {
  // Try to extract trace context from caller
  const activeContext = opentelemetry.propagation.extract(
    opentelemetry.context.active(),
    request.meta?.tracing
  );

  // Run the whole request with the extracted context
  return opentelemetry.context.with(activeContext, () => {
    // Wrapping span for the whole request
    return withSpan(request.method, async (span) => {
      let db = null;

      try {
        const { createJobContextAPI, jobs } = config;

        if (!(request.method in jobs)) {
          const message = `no corresponding job found for '${request.method}'`;
          span.setStatus({
            code: opentelemetry.SpanStatusCode.ERROR,
            message: message,
          });
          return createJSONRPCErrorResponse(
            request.id,
            JSONRPCErrorCode.MethodNotFound,
            message
          );
        }

        // The ctx argument passed into the job function.
        const ctx = createJobContextAPI({
          meta: request.meta,
        });

        const permitted =
          request.meta && request.meta.permissionState.status === "granted"
            ? true
            : null;

        db = getDatabaseClient();
        const jobFunction = jobs[request.method];
        const actionType = PROTO_ACTION_TYPES.JOB;

        await tryExecuteJob(
          { request, permitted, db, actionType },
          async () => {
            // parse request params to convert objects into rich field types (e.g. InlineFile)
            const inputs = parseParams(request.params);

            // Return the job function to the containing tryExecuteJob block
            return jobFunction(ctx, inputs);
          }
        );

        return createJSONRPCSuccessResponse(request.id, null);
      } catch (e) {
        if (e instanceof Error) {
          span.recordException(e);
          span.setStatus({
            code: opentelemetry.SpanStatusCode.ERROR,
            message: e.message,
          });
          return errorToJSONRPCResponse(request, e);
        }

        const message = JSON.stringify(e);

        span.setStatus({
          code: opentelemetry.SpanStatusCode.ERROR,
          message: message,
        });

        return createJSONRPCErrorResponse(
          request.id,
          RuntimeErrors.UnknownError,
          message
        );
      } finally {
        if (db != null) {
          await db.destroy();
        }
      }
    });
  });
}

module.exports = {
  handleJob,
  RuntimeErrors,
};
