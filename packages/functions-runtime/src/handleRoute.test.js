import { createJSONRPCRequest, JSONRPCErrorCode } from "json-rpc-2.0";
import { handleRoute, RuntimeErrors } from "./handleRoute";
import { test, expect } from "vitest";

test("when a route function does not exist or has not been implemented", async () => {
  const config = {
    functions: {
      myRoute: async (params, ctx) => {
        return {
          body: "Hello World",
          headers: {},
        };
      },
    },
    createContextAPI: () => {
      return {
        response: {
          headers: new Headers(),
        },
      };
    },
  };

  const rpcReq = createJSONRPCRequest("123", "nonExistentRoute", {});

  expect(await handleRoute(rpcReq, config)).toEqual({
    id: "123",
    jsonrpc: "2.0",
    error: {
      code: JSONRPCErrorCode.MethodNotFound,
      message:
        "route function 'nonExistentRoute' does not exist or has not been implemented",
    },
  });
});
