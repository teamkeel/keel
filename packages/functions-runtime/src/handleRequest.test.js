import { createJSONRPCRequest, JSONRPCErrorCode } from "json-rpc-2.0";
import { handleRequest, RuntimeErrors } from "./handleRequest";
import { test, expect, vi } from "vitest";
const { Permissions } = require("./permissions");
import { PROTO_ACTION_TYPES } from "./consts";

process.env.KEEL_DB_CONN_TYPE = "pg";
process.env.KEEL_DB_CONN = `postgresql://postgres:postgres@localhost:5432/functions-runtime`;

test("when the custom function returns expected value", async () => {
  const config = {
    functions: {
      createPost: async (ctx, inputs) => {
        new Permissions().allow();

        return {
          title: "a post",
          id: "abcde",
        };
      },
    },
    actionTypes: {
      createPost: PROTO_ACTION_TYPES.CREATE,
    },
    createContextAPI: () => {},
  };

  const rpcReq = newRpcRequest("123", "createPost", { title: "a post" });

  expect(await handleRequest(rpcReq, config)).toEqual({
    id: "123",
    jsonrpc: "2.0",
    meta: {
      headers: {},
    },
    result: {
      title: "a post",
      id: "abcde",
    },
  });
});

test("when the custom function doesnt return a value", async () => {
  const config = {
    functions: {
      createPost: async (ctx, inputs) => {
        new Permissions().allow();
      },
    },
    permissions: {},
    actionTypes: {
      createPost: PROTO_ACTION_TYPES.CREATE,
    },
    createContextAPI: () => {},
  };

  const rpcReq = newRpcRequest("123", "createPost", { title: "a post" });

  expect(await handleRequest(rpcReq, config)).toEqual({
    id: "123",
    jsonrpc: "2.0",
    error: {
      code: RuntimeErrors.NoResultError,
      message: "no result returned from function 'createPost'",
    },
  });
});

test("when there is no matching function for the path", async () => {
  const config = {
    functions: {
      createPost: async (ctx, inputs) => {},
    },
    actionTypes: {
      createPost: PROTO_ACTION_TYPES.CREATE,
    },
    createContextAPI: () => {},
  };

  const rpcReq = newRpcRequest("123", "unknown", { title: "a post" });

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
      createPost: async (ctx, inputs) => {
        throw new Error("oopsie daisy");
      },
    },
    actionTypes: {
      createPost: PROTO_ACTION_TYPES.CREATE,
    },
    createContextAPI: () => {},
  };

  const rpcReq = newRpcRequest("123", "createPost", { title: "a post" });

  expect(await handleRequest(rpcReq, config)).toEqual({
    id: "123",
    jsonrpc: "2.0",
    error: {
      code: RuntimeErrors.UnknownError,
      message: "oopsie daisy",
    },
  });
});

test("when a role based permission has already been granted by the main runtime", async () => {
  const config = {
    functions: {
      createPost: async (ctx, inputs, api) => {
        return {
          title: inputs.title,
        };
      },
    },
    actionTypes: {
      createPost: PROTO_ACTION_TYPES.CREATE,
    },
    createContextAPI: () => {},
  };

  let rpcReq = newRpcRequest(
    "123",
    "createPost",
    { title: "a post" },
    { permissionState: { status: "granted", reason: "role" } }
  );

  expect(await handleRequest(rpcReq, config)).toEqual({
    id: "123",
    jsonrpc: "2.0",
    result: {
      title: "a post",
    },
    meta: {
      headers: {},
    },
  });
});

test("when there is an unexpected object thrown in the custom function", async () => {
  const config = {
    functions: {
      createPost: async (ctx, inputs) => {
        throw { err: "oopsie daisy" };
      },
    },
    actionTypes: {
      createPost: PROTO_ACTION_TYPES.CREATE,
    },
    createContextAPI: () => {},
  };

  const rpcReq = newRpcRequest("123", "createPost", { title: "a post" });

  expect(await handleRequest(rpcReq, config)).toEqual({
    id: "123",
    jsonrpc: "2.0",
    error: {
      code: RuntimeErrors.UnknownError,
      message: '{"err":"oopsie daisy"}',
    },
  });
});

function newRpcRequest(
  id,
  method,
  params,
  meta = { permissionState: { status: "unknown" } }
) {
  const req = createJSONRPCRequest(id, method, params);

  Object.assign(req, {
    ...req,
    meta,
  });

  return req;
}
