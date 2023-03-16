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
const { PROTO_ACTION_TYPES_REQUEST_HANDLER } = require("./consts");

const { errorToJSONRPCResponse, RuntimeErrors } = require("./errors");

// Generic handler function that is agnostic to runtime environment (local or lambda)
// to execute a custom function based on the contents of a jsonrpc-2.0 payload object.
// To read more about jsonrpc request and response shapes, please read https://www.jsonrpc.org/specification
async function handleRequest(request, config) {
  const { createFunctionAPI, createContextAPI, functions, permissions } =
    config;

  if (!(request.method in functions)) {
    return createJSONRPCErrorResponse(
      request.id,
      JSONRPCErrorCode.MethodNotFound,
      `no corresponding function found for '${request.method}'`
    );
  }

  try {
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
        headers,
        db: transaction,
      });

      const customFunction = functions[request.method];
      // Call the user's custom function!
      const fnResult = await customFunction.fn(request.params, api, ctx);

      // api.permissions maintains an internal state of whether the current operation has been *explicitly* permitted/denied by the user in the course of their custom function.
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

          // We only want to run permission checks at the handleRequest level for action types list, get and create
          // Delete and update permission checks need to be baked into the model api because they require reading records to be deleted / updated from the database first in order to ascertain whether the records to be deleted or updated fulfil the permission
          if (
            PROTO_ACTION_TYPES_REQUEST_HANDLER.includes(
              customFunction.actionType
            )
          ) {
            // check will throw a PermissionError if a permission rule is invalid
            await checkBuiltInPermissions({
              rows: fnResult,
              permissions: relevantPermissions,
              db: transaction,
              ctx,
              functionName: request.method,
            });
          }

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
      return errorToJSONRPCResponse(request, e);
    }

    return createJSONRPCErrorResponse(
      request.id,
      RuntimeErrors.UnknownError,
      JSON.stringify(e)
    );
  }
}

module.exports = {
  handleRequest,
  RuntimeErrors,
};
