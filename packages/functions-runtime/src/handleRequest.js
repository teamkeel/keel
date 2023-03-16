const {
  createJSONRPCErrorResponse,
  createJSONRPCSuccessResponse,
  JSONRPCErrorCode,
} = require("json-rpc-2.0");

const { getDatabase } = require("./database");
const { PERMISSION_STATE, PermitError } = require("./permissions");

const { errorToJSONRPCResponse, RuntimeErrors } = require("./errors");

// Generic handler function that is agnostic to runtime environment (local or lambda)
// to execute a custom function based on the contents of a jsonrpc-2.0 payload object.
// To read more about jsonrpc request and response shapes, please read https://www.jsonrpc.org/specification
async function handleRequest(request, config) {
  const { createFunctionAPI, createContextAPI, functions } = config;

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
    const result = await db.transaction().execute(async (trx) => {
      const api = createFunctionAPI({ headers, db: trx });
      const ctx = createContextAPI(request.meta);

      // Call the user's custom function!
      const fnResult = await functions[request.method](
        request.params,
        api,
        ctx
      );

      // api.permissions maintains an internal state of whether the current operation has been permitted either by the user or by built-in permission rules
      // we need to check that the final state is permitted. if it's not, then we want to rollback
      // the transaction
      if (api.permissions.getState() !== PERMISSION_STATE.PERMITTED) {
        // Any error thrown inside of Kysely's transaction execute() will cause the transaction to be rolled back.
        // PermitError is handled by our JSONRPC error serialisation code
        throw new PermitError(`Not permitted to access ${request.method}`);
      } else {
        // otherwise, if everything is permitted, then we just return the function result from
        // the transaction closure.
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
