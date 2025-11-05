import { NumberInputWithActions, FlowConfig } from "@teamkeel/sdk";

const config = {
  // See https://docs.keel.so/flows for options
} as const satisfies FlowConfig;

export default NumberInputWithActions(config, async (ctx) => {
  const page1 = await ctx.ui.page("number input validation", {
    content: [
      ctx.ui.inputs.number("quantity", {
        label: "Quantity",
        validate: (data, action) => {
          // Verify action parameter is passed correctly when an action is provided
          if (action !== undefined && action !== "buy" && action !== "reserve") {
            throw new Error(
              `Expected action to be 'buy' or 'reserve', got: ${action}`
            );
          }

          // Different validation rules per action
          if (action === "buy" && data < 1) {
            return "Must buy at least 1 item";
          }
          if (action === "reserve" && data < 5) {
            return "Must reserve at least 5 items";
          }
          return true;
        },
      }),
    ],
    actions: [
      { label: "Buy", value: "buy" },
      { label: "Reserve", value: "reserve" },
    ],
  });

  return {
    action: page1.action,
    quantity: page1.data.quantity,
  };
});
