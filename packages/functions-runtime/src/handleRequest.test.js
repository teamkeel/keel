import { createJSONRPCRequest, JSONRPCErrorCode } from "json-rpc-2.0";
import { handleRequest, RuntimeErrors } from "./handleRequest";
import { test, expect } from "vitest";

test("when the custom function returns expected value", async () => {
  const config = {
    functions: {
      createPost: async () => {
        return {
          title: "a post",
          id: "abcde",
        };
      },
    },
    createFunctionAPI: () => {},
  };

  const rpcReq = createJSONRPCRequest("123", "createPost", { title: "a post" });

  expect(await handleRequest(rpcReq, config)).toEqual({
    id: "123",
    jsonrpc: "2.0",
    result: {
      title: "a post",
      id: "abcde",
    },
  });
});

test("when the custom function doesnt return a value", async () => {
  const config = {
    functions: {
      createPost: async () => {},
    },
    createFunctionAPI: () => {},
  };

  const rpcReq = createJSONRPCRequest("123", "createPost", { title: "a post" });

  expect(await handleRequest(rpcReq, config)).toEqual({
    id: "123",
    jsonrpc: "2.0",
    error: {
      code: JSONRPCErrorCode.InternalError,
      message: "no result returned from function 'createPost'",
    },
  });
});

test("when there is no matching function for the path", async () => {
  const config = {
    functions: {
      createPost: async () => {},
    },
    createFunctionAPI: () => {},
  };

  const rpcReq = createJSONRPCRequest("123", "unknown", { title: "a post" });

  expect(await handleRequest(rpcReq, config)).toEqual({
    id: "123",
    jsonrpc: "2.0",
    error: {
      code: JSONRPCErrorCode.MethodNotFound,
      message: "no corresponding function found for 'unknown'",
    },
  });
});

test("when there is an unexpected error in the custom function", async () => {
  const config = {
    functions: {
      createPost: () => {
        throw new Error("oopsie daisy");
      },
    },
    createFunctionAPI: () => {},
  };

  const rpcReq = createJSONRPCRequest("123", "createPost", { title: "a post" });

  expect(await handleRequest(rpcReq, config)).toEqual({
    id: "123",
    jsonrpc: "2.0",
    error: {
      code: RuntimeErrors.UnknownError,
      message: "oopsie daisy",
      data: {
        stack: expect.stringContaining('Error: oopsie daisy')
      }
    },
  });
});

test("when there is an unexpected object thrown in the custom function", async () => {
  const config = {
    functions: {
      createPost: () => {
        throw { err: "oopsie daisy" };
      },
    },
    createFunctionAPI: () => {},
  };

  const rpcReq = createJSONRPCRequest("123", "createPost", { title: "a post" });

  expect(await handleRequest(rpcReq, config)).toEqual({
    id: "123",
    jsonrpc: "2.0",
    error: {
      code: RuntimeErrors.UnknownError,
      message: '{"err":"oopsie daisy"}',
    },
  });
});
