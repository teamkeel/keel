import { IteratorWithActions, FlowConfig } from "@teamkeel/sdk";

const config = {
  // See https://docs.keel.so/flows for options
} as const satisfies FlowConfig;

export default IteratorWithActions(config, async (ctx) => {
  const page1 = await ctx.ui.page("iterator validation", {
    content: [
      ctx.ui.iterator("items", {
        content: [
          ctx.ui.inputs.text("itemName", { label: "Item Name" }),
          ctx.ui.inputs.number("price", {
            label: "Price",
            validate: (data, action) => {
              // Verify action parameter is passed correctly when an action is provided
              if (action !== undefined && action !== "finalize" && action !== "save") {
                throw new Error(
                  `Expected action to be 'finalize' or 'save', got: ${action}`
                );
              }

              // Prices must be positive when finalizing
              if (action === "finalize" && data <= 0) {
                return "Price must be positive when finalizing";
              }
              return true;
            },
          }),
        ],
        validate: (data, action) => {
          // Verify action parameter is passed correctly when an action is provided
          if (action !== undefined && action !== "finalize" && action !== "save") {
            throw new Error(
              `Expected action to be 'finalize' or 'save', got: ${action}`
            );
          }

          // Must have at least 2 items when finalizing
          if (action === "finalize" && data.length < 2) {
            return "Must have at least 2 items to finalize";
          }
          return true;
        },
      }),
    ],
    actions: ["finalize", "save"],
  });

  return {
    action: page1.action,
    items: page1.data["items"],
  };
});
