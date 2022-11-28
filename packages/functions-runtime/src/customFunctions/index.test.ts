import { createJSONRPCRequest, JSONRPCErrorCode } from "json-rpc-2.0";
import { Config } from "../types";
import handle from ".";

test("when the custom function returns expected value", async () => {
  const config: Config = {
    functions: {
      createPost: () => {
        return {
          title: "a post",
          id: "abcde",
        };
      },
    },
    api: {},
  };

  const rpcReq = createJSONRPCRequest("123", "createPost", { title: "a post" });

  expect(await handle(rpcReq, config)).toEqual({
    id: "123",
    jsonrpc: "2.0",
    result: {
      title: "a post",
      id: "abcde",
    },
  });
});

test("when the custom function doesnt return a value", async () => {
  const config: Config = {
    functions: {
      createPost: () => {},
    },
    api: {},
  };

  const rpcReq = createJSONRPCRequest("123", "createPost", { title: "a post" });

  expect(await handle(rpcReq, config)).toEqual({
    id: "123",
    jsonrpc: "2.0",
    error: {
      code: JSONRPCErrorCode.InternalError,
      message: "no result returned from function 'createPost'",
    },
  });
});

test("when there is no matching function for the path", async () => {
  const config: Config = {
    functions: {
      createPost: () => {},
    },
    api: {},
  };

  const rpcReq = createJSONRPCRequest("123", "unknown", { title: "a post" });

  expect(await handle(rpcReq, config)).toEqual({
    id: "123",
    jsonrpc: "2.0",
    error: {
      code: JSONRPCErrorCode.InvalidRequest,
      message: "no corresponding function found for 'unknown'",
    },
  });
});
