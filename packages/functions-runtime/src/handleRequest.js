const {
  createJSONRPCErrorResponse,
  createJSONRPCSuccessResponse,
  JSONRPCErrorCode,
} = require("json-rpc-2.0");
const { getDatabaseClient } = require("./database");
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

        // The Go runtime does *some* permissions checks up front before the request reaches
        // this method, so we pass in a permissionState object on the request.meta object that
        // indicates whether a call to a custom function has already been authorised
        const permitted =
          request.meta && request.meta.permissionState.status === "granted"
            ? true
            : null;

        const db = getDatabaseClient();
        const customFunction = functions[request.method];
        const actionType = actionTypes[request.method];

        const result = await tryExecuteFunction(
          { request, ctx, permitted, db, permissionFns, actionType },
          async () => {
            // Return the custom function to the containing tryExecuteFunction block
            // Once the custom function is called, tryExecuteFunction will check the schema's permission rules to see if it can continue committing
            // the transaction to the db. If a permission rule is violated, any changes made inside the transaction are rolled back.
            return customFunction(ctx, request.params);
          }
        );

        const response = createJSONRPCSuccessResponse(request.id, result);

        const responseHeaders = {};
        for (const pair of headers.entries()) {
          responseHeaders[pair[0]] = pair[1].split(", ");
        }
        response.meta = {
          headers: responseHeaders,
          status: ctx.response.status,
        };

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
