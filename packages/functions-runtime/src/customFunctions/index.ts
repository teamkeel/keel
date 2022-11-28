import {
  createJSONRPCErrorResponse,
  createJSONRPCSuccessResponse,
  JSONRPCErrorCode,
} from "json-rpc-2.0";

import {
  Config,
  CustomFunctionResponsePayload,
  CustomFunctionRequestPayload,
} from "../types";

// Generic handler function that is agnostic to runtime environment (http or lambda)
// to execute a custom function based on the contents of a jsonrpc-2.0 payload object.
// To read more about jsonrpc request and response shapes, please read https://www.jsonrpc.org/specification
const handler = async (
  { method: name, params, id }: CustomFunctionRequestPayload,
  config: Config
): Promise<CustomFunctionResponsePayload> => {
  const { api, functions } = config;

  if (!(name in functions)) {
    return createJSONRPCErrorResponse(
      id,
      JSONRPCErrorCode.MethodNotFound,
      `no corresponding function found for '${name}'`
    );
  }

  try {
    const result = await functions[name].call(params, api);

    if (!result) {
      // no result returned from custom function
      return createJSONRPCErrorResponse(
        id,
        JSONRPCErrorCode.ParseError,
        `no result returned from function '${name}'`
      );
    }

    return createJSONRPCSuccessResponse(id, result);
  } catch (e) {
    let msg = "";

    if (e instanceof Error) {
      msg = e.message;
    } else {
      msg = JSON.stringify(e);
    }

    return createJSONRPCErrorResponse(
      id,
      JSONRPCErrorCode.InternalError,
      e.message
    );
  }
};

export default handler;
