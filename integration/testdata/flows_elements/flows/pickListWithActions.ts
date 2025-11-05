import { PickListWithActions, FlowConfig } from "@teamkeel/sdk";

const config = {
  // See https://docs.keel.so/flows for options
} as const satisfies FlowConfig;

export default PickListWithActions(config, async (ctx) => {
  const products = [
    {
      id: "prod-1",
      name: "Widget A",
      targetQty: 10,
      barcodes: ["1234567890"],
    },
    {
      id: "prod-2",
      name: "Widget B",
      targetQty: 5,
      barcodes: ["0987654321"],
    },
    {
      id: "prod-3",
      name: "Widget C",
      targetQty: 3,
      barcodes: ["1111111111"],
    },
  ];

  const page1 = await ctx.ui.page("pick list with actions", {
    content: [
      ctx.ui.interactive.pickList("items", {
        data: products,
        render: (product) => ({
          id: product.id,
          targetQuantity: product.targetQty,
          title: product.name,
          barcodes: product.barcodes,
        }),
        validate: (response, action) => {
          // Verify action parameter is passed correctly when an action is provided
          if (
            action !== undefined &&
            action !== "complete" &&
            action !== "partial" &&
            action !== "cancel"
          ) {
            throw new Error(
              `Expected action to be 'complete', 'partial', or 'cancel', got: ${action}`
            );
          }

          // Different validation rules based on action
          if (action === "complete") {
            // When completing, all items must be fully picked
            const allComplete = response.items.every(
              (item) => item.quantity >= item.targetQuantity
            );
            if (!allComplete) {
              return "All items must be fully picked to complete";
            }

            // Total quantity must match total target
            const totalPicked = response.items.reduce(
              (sum, item) => sum + item.quantity,
              0
            );
            const totalTarget = response.items.reduce(
              (sum, item) => sum + item.targetQuantity,
              0
            );
            if (totalPicked !== totalTarget) {
              return "Total picked must match total target quantity";
            }
          }

          if (action === "partial") {
            // Partial allows any quantity, but at least one item must be picked
            const anyPicked = response.items.some((item) => item.quantity > 0);
            if (!anyPicked) {
              return "At least one item must be picked for partial completion";
            }
          }

          if (action === "cancel") {
            // Cancel allows any state
            return true;
          }

          return true;
        },
      }),
    ],
    actions: ["complete", "partial", "cancel"],
  });

  return {
    action: page1.action,
    items: page1.data.items,
  };
});
