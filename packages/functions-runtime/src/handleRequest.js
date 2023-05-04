const {
  createJSONRPCErrorResponse,
  createJSONRPCSuccessResponse,
  JSONRPCErrorCode,
} = require("json-rpc-2.0");

const { getDatabase } = require("./database");
const {
  PERMISSION_STATE,
  PermissionError,
  checkBuiltInPermissions,
} = require("./permissions");
const { PROTO_ACTION_TYPES } = require("./consts");

const { errorToJSONRPCResponse, RuntimeErrors } = require("./errors");
const opentelemetry = require("@opentelemetry/api");
const { serviceName } = require("./tracing");

// Generic handler function that is agnostic to runtime environment (local or lambda)
// to execute a custom function based on the contents of a jsonrpc-2.0 payload object.
// To read more about jsonrpc request and response shapes, please read https://www.jsonrpc.org/specification
async function handleRequest(request, config) {
  const activeContext = opentelemetry.propagation.extract(
    opentelemetry.context.active(),
    // "?." is so we don't have to provide this field on tests
    request.meta?.tracing
  );
  const tracer = opentelemetry.trace.getTracer(serviceName);
  let span = tracer.startSpan(
    `Function/${request.method}`,
    { attributes: {} },
    activeContext
  );
  opentelemetry.trace.setSpan(activeContext, span);

  //
  // WARNING: Nothing should be done before this try block, as the finally closes the span
  //
  try {
    const {
      createFunctionAPI,
      createContextAPI,
      functions,
      permissions,
      actionTypes,
    } = config;

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

    const db = getDatabase();

    // We want to wrap the execution of the custom function in a transaction so that any call the user makes
    // to any of the model apis we provide to the custom function is processed in a transaction.
    // This is useful for permissions where we want to only proceed with database writes if all permission rules
    // have been validated.
    const result = await db.transaction().execute(async (transaction) => {
      const ctx = createContextAPI(request.meta);
      const api = createFunctionAPI({
        meta: request.meta,
        headers,
        db: transaction,
      });

      const customFunction = functions[request.method];

      // Call the user's custom function!
      const fnResult = await customFunction(request.params, api, ctx);

      // api.permissions maintains an internal state of whether the current operation has been *explicitly* permitted/denied by the user in the course of their custom function, or if execution has already been permitted by a role based permission (evaluated in the main runtime).
      // we need to check that the final state is permitted or unpermitted. if it's not, then it means that the user has taken no explicit action to permit/deny
      // and therefore we default to checking the permissions defined in the schema automatically.
      switch (api.permissions.getState()) {
        case PERMISSION_STATE.PERMITTED:
          return fnResult;
        case PERMISSION_STATE.UNPERMITTED:
          throw new PermissionError(
            `Not permitted to access ${request.method}`
          );
        default:
          // unknown state, proceed with checking against the built in permissions in the schema
          const relevantPermissions = permissions[request.method];

          const actionType = actionTypes[request.method];

          const peakInsideTransaction =
            actionType === PROTO_ACTION_TYPES.CREATE;

          let rowsForPermissions = [];
          switch (actionType) {
            case PROTO_ACTION_TYPES.LIST:
              rowsForPermissions = fnResult;

              break;
            case PROTO_ACTION_TYPES.DELETE:
              rowsForPermissions = [{ id: fnResult }];
              break;
            default:
              rowsForPermissions = [fnResult];
              break;
          }

          // check will throw a PermissionError if a permission rule is invalid
          await checkBuiltInPermissions({
            rows: rowsForPermissions,
            permissions: relevantPermissions,
            // it is important that we pass db here as db represents the connection to the database
            // *outside* of the current transaction. Given that any changes inside of a transaction
            // are opaque to the outside, we can utilize this when running permission rules and then deciding to
            // rollback any changes if they do not pass. However, for creates we need to be able to 'peak' inside the transaction to read the created record, as this won't exist outside of the transaction.
            db: peakInsideTransaction ? transaction : db,
            ctx,
            functionName: request.method,
          });

          // If the built in permission check above doesn't throw, then it means that the request is permitted and we can continue returning the return value from the custom function out of the transaction
          return fnResult;
      }
    });

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
  } finally {
    span.end();
  }
}

module.exports = {
  handleRequest,
  RuntimeErrors,
};
