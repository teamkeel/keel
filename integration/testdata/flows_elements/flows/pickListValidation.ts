import { PickListValidation, FlowConfig } from "@teamkeel/sdk";

const config = {
  // See https://docs.keel.so/flows for options
} as const satisfies FlowConfig;

export default PickListValidation(config, async (ctx) => {
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

  const page1 = await ctx.ui.page("pick list page", {
    content: [
      ctx.ui.interactive.pickList("items", {
        data: products,
        render: (product) => ({
          id: product.id,
          targetQuantity: product.targetQty,
          title: product.name,
          barcodes: product.barcodes,
        }),
        validate: (response) => {
          // Validate that total quantity doesn't exceed 20
          const totalQuantity = response.items.reduce(
            (sum, item) => sum + item.quantity,
            0
          );
          if (totalQuantity > 20) {
            return "Total quantity cannot exceed 20 items";
          }

          // Validate that at least one item is picked
          if (response.items.every((item) => item.quantity === 0)) {
            return "At least one item must be picked";
          }

          return true;
        },
      }),
    ],
  });

  // Verify the items were picked correctly
  if (page1.items.items.length !== 3) {
    throw new Error("Expected 3 items in response");
  }

  return page1.items;
});
