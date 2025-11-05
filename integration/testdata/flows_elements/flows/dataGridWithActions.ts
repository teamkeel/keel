import { DataGridWithActions, FlowConfig } from "@teamkeel/sdk";

const config = {
  // See https://docs.keel.so/flows for options
} as const satisfies FlowConfig;

export default DataGridWithActions(config, async (ctx) => {
  // Test dataGrid validation with actions
  const page1 = await ctx.ui.page("inventory management", {
    content: [
      ctx.ui.inputs.dataGrid("inventory", {
        data: [
          { id: "item-1", sku: "SKU001", quantity: 10, price: 99.99 },
          { id: "item-2", sku: "SKU002", quantity: 5, price: 149.99 },
        ],
        columns: [
          { key: "id", label: "ID", type: "id", editable: false },
          { key: "sku", label: "SKU", type: "text", editable: true },
          { key: "quantity", label: "Qty", type: "number", editable: true },
          { key: "price", label: "Price", type: "number", editable: true },
        ],
        allowAddRows: true,
        allowDeleteRows: true,
        validate: (data, action) => {
          // Verify action parameter is passed correctly when an action is provided
          if (
            action !== undefined &&
            action !== "approve" &&
            action !== "draft"
          ) {
            throw new Error(
              `Expected action to be 'approve' or 'draft', got: ${action}`
            );
          }

          // Different validation rules based on action
          if (action === "approve") {
            // When approving, all items must have positive quantities
            const hasZeroQty = data.some((item) => item.quantity <= 0);
            if (hasZeroQty) {
              return "All items must have positive quantities when approving";
            }

            // When approving, must have at least 2 items
            if (data.length < 2) {
              return "Must have at least 2 items to approve";
            }

            // Total value must be at least $100
            const totalValue = data.reduce(
              (sum, item) => sum + item.quantity * item.price,
              0
            );
            if (totalValue < 100) {
              return "Total value must be at least $100 to approve";
            }
          }

          if (action === "draft") {
            // Draft allows empty data
            return true;
          }

          // Common validation for all actions
          const hasNegativePrice = data.some((item) => item.price < 0);
          if (hasNegativePrice) {
            return "Prices cannot be negative";
          }

          return true;
        },
      }),
    ],
    actions: ["approve", "draft"],
  });

  return {
    action: page1.action,
    inventory: page1.data.inventory,
  };
});
