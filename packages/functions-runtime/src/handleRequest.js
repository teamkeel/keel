const {
  createJSONRPCErrorResponse,
  createJSONRPCSuccessResponse,
  JSONRPCErrorCode,
} = require("json-rpc-2.0");
const { getDatabase } = require("./database");
const { tryExecuteFunction } = require("./tryExecuteFunction");
const { errorToJSONRPCResponse, RuntimeErrors } = require("./errors");
const opentelemetry = require("@opentelemetry/api");
const { withSpan } = require("./tracing");

// Generic handler function that is agnostic to runtime environment (local or lambda)
// to execute a custom function based on the contents of a jsonrpc-2.0 payload object.
// To read more about jsonrpc request and response shapes, please read https://www.jsonrpc.org/specification
async function handleRequest(request, config) {
  // Try to extract trace context from caller
  const activeContext = opentelemetry.propagation.extract(
    opentelemetry.context.active(),
    request.meta?.tracing
  );

  // Run the whole request with the extracted context
  return opentelemetry.context.with(activeContext, () => {
    // Wrapping span for the whole request
    return withSpan(request.method, async (span) => {
      try {
        const { createContextAPI, functions, permissionFns, actionTypes } =
          config;

        if (!(request.method in functions)) {
          const message = `no corresponding function found for '${request.method}'`;
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

        // headers reference passed to custom function where object data can be modified
        const headers = new Headers();

        // The ctx argument passed into the custom function.
        const ctx = createContextAPI({
          responseHeaders: headers,
          meta: request.meta,
        });

        const permitted =
          request.meta && request.meta.permissionState.status === "granted"
            ? true
            : null;

        const db = getDatabase();
        const customFunction = functions[request.method];

        const result = await tryExecuteFunction(
          { request, ctx, permitted, db, permissionFns, actionTypes },
          async () => {
            // Call the user's custom function!
            return customFunction(ctx, request.params);
          }
        );

        // We want to wrap the execution of the custom function in a transaction so that any call the user makes
        // to any of the model apis we provide to the custom function is processed in a transaction.
        // This is useful for permissions where we want to only proceed with database writes if all permission rules
        // have been validated.
        if (result === undefined) {
          // no result returned from custom function
          return createJSONRPCErrorResponse(
            request.id,
            RuntimeErrors.NoResultError,
            `no result returned from function '${request.method}'`
          );
        }

        const response = createJSONRPCSuccessResponse(request.id, result);

        const responseHeaders = {};
        for (const pair of headers.entries()) {
          responseHeaders[pair[0]] = pair[1].split(", ");
        }
        response.meta = { headers: responseHeaders };

        return response;
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
      }
    });
  });
}

module.exports = {
  handleRequest,
  RuntimeErrors,
};
