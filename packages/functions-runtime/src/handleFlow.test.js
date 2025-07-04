import { createJSONRPCRequest, JSONRPCErrorCode } from "json-rpc-2.0";
import { handleFlow, RuntimeErrors } from "./handleFlow";
import { test, expect } from "vitest";

test("when a flow does not exist or has not been implemented", async () => {
  const config = {
    flows: {
      myFlow: {
        fn: async (ctx, inputs) => {},
        config: {
          title: "My Flow",
          stages: ["step1", "step2"],
        },
      },
    },
    createFlowContextAPI: () => ({}),
  };

  const rpcReq = createJSONRPCRequest("123", "nonExistentFlow", {});
  rpcReq.meta = {
    runId: "test-run-id",
    data: {},
    action: "test-action",
  };

  expect(await handleFlow(rpcReq, config)).toEqual({
    id: "123",
    jsonrpc: "2.0",
    error: {
      code: JSONRPCErrorCode.MethodNotFound,
      message:
        "flow 'nonExistentFlow' does not exist or has not been implemented",
    },
  });
});

test("when no runId is provided", async () => {
  const config = {
    flows: {
      myFlow: {
        fn: async (ctx, inputs) => {},
        config: {
          title: "My Flow",
          stages: ["step1", "step2"],
        },
      },
    },
    createFlowContextAPI: () => ({}),
  };

  const rpcReq = createJSONRPCRequest("123", "myFlow", {});
  // No runId in meta

  expect(await handleFlow(rpcReq, config)).toEqual({
    id: "123",
    jsonrpc: "2.0",
    error: {
      code: RuntimeErrors.UnknownError,
      message: "no runId provided",
    },
  });
});
