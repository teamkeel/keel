const {
  createJSONRPCErrorResponse,
  createJSONRPCSuccessResponse,
  JSONRPCErrorCode,
} = require("json-rpc-2.0");
const { createDatabaseClient } = require("./database");
const { errorToJSONRPCResponse, RuntimeErrors } = require("./errors");
const opentelemetry = require("@opentelemetry/api");
const { withSpan } = require("./tracing");
const { tryExecuteFlow } = require("./tryExecuteFlow");
const { parseInputs } = require("./parsing");

async function handleFlow(request, config) {
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
        const { createFlowContextAPI, flows } = config;

        if (!(request.method in flows)) {
          const message = `no corresponding flow found for '${request.method}'`;
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

        // The ctx argument passed into the flow function.
        const ctx = createFlowContextAPI();

        db = createDatabaseClient({
          connString: request.meta?.secrets?.KEEL_DB_CONN,
        });

        const flowFunction = flows[request.method];

        await tryExecuteFlow({ request, db }, async () => {
          // parse request params to convert objects into rich field types (e.g. InlineFile)
          const inputs = parseInputs(request.params);

          // Return the job function to the containing tryExecuteJob block
          return flowFunction(ctx, inputs);
        });

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
        if (db) {
          await db.destroy();
        }
      }
    });
  });
}

module.exports = {
  handleFlow,
  RuntimeErrors,
};
