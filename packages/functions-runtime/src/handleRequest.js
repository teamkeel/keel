const {
  createJSONRPCErrorResponse,
  createJSONRPCSuccessResponse,
  JSONRPCErrorCode,
} = require("json-rpc-2.0");

// Generic handler function that is agnostic to runtime environment (local or lambda)
// to execute a custom function based on the contents of a jsonrpc-2.0 payload object.
// To read more about jsonrpc request and response shapes, please read https://www.jsonrpc.org/specification
export async function handleRequest(request, config) {
  const { createFunctionAPI, functions } = config;

  if (!(request.method in functions)) {
    return createJSONRPCErrorResponse(
      request.id,
      JSONRPCErrorCode.MethodNotFound,
      `no corresponding function found for '${request.method}'`
    );
  }

  try {
    const result = await functions[request.method](
      request.params,
      createFunctionAPI()
    );

    if (result === undefined) {
      // no result returned from custom function
      return createJSONRPCErrorResponse(
        request.id,
        JSONRPCErrorCode.InternalError,
        `no result returned from function '${request.method}'`
      );
    }

    return createJSONRPCSuccessResponse(request.id, result);
  } catch (e) {
    let msg = "";

    if (e instanceof Error) {
      msg = e.message;
    } else {
      msg = JSON.stringify(e);
    }

    return createJSONRPCErrorResponse(
      request.id,
      JSONRPCErrorCode.InternalError,
      msg
    );
  }
}
