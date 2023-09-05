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
const { tryExecuteSubscriber } = require("./tryExecuteSubscriber");

// Generic handler function that is agnostic to runtime environment (local or lambda)
// to execute a subscriber function based on the contents of a jsonrpc-2.0 payload object.
// To read more about jsonrpc request and response shapes, please read https://www.jsonrpc.org/specification
async function handleSubscriber(request, config) {
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
        const { createSubscriberContextAPI, subscribers } = config;

        if (!(request.method in subscribers)) {
          const message = `no corresponding subscriber found for '${request.method}'`;
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

        // The ctx argument passed into the subscriber function.
        const ctx = createSubscriberContextAPI({
          meta: request.meta,
        });

        const db = getDatabaseClient();
        const subscriberFunction = subscribers[request.method];
        const actionType = PROTO_ACTION_TYPES.SUBSCRIBER;

        await tryExecuteSubscriber({ request, db, actionType }, async () => {
          // Return the subscriber function to the containing tryExecuteSubscriber block
          return subscriberFunction(ctx, request.params);
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
      }
    });
  });
}

module.exports = {
  handleSubscriber,
  RuntimeErrors,
};
