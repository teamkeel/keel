import { createJSONRPCRequest, JSONRPCErrorCode } from "json-rpc-2.0";
import { handleSubscriber, RuntimeErrors } from "./handleSubscriber";
import { test, expect } from "vitest";

test("when a subscriber does not exist or has not been implemented", async () => {
  const config = {
    subscribers: {
      mySubscriber: async (ctx, inputs) => {
        // Handle the event
      },
    },
    createSubscriberContextAPI: () => ({}),
  };

  const rpcReq = createJSONRPCRequest("123", "nonExistentSubscriber", {});

  expect(await handleSubscriber(rpcReq, config)).toEqual({
    id: "123",
    jsonrpc: "2.0",
    error: {
      code: JSONRPCErrorCode.MethodNotFound,
      message:
        "subscriber 'nonExistentSubscriber' does not exist or has not been implemented",
    },
  });
});
