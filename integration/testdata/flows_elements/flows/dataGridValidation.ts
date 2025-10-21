import { DataGridValidation, FlowConfig } from "@teamkeel/sdk";

const config = {
  // See https://docs.keel.so/flows for options
} as const satisfies FlowConfig;

export default DataGridValidation(config, async (ctx) => {
  // Test 1: Basic dataGrid with inferred columns
  await ctx.ui.page("basic data grid", {
    content: [
      ctx.ui.inputs.dataGrid("products", {
        data: [
          { id: "prod-1", name: "Widget A", quantity: 10, inStock: true },
          { id: "prod-2", name: "Widget B", quantity: 5, inStock: false },
          { id: "prod-3", name: "Widget C", quantity: 0, inStock: true },
        ],
      }),
    ],
  });

  // Test 2: DataGrid with explicit columns and validation
  await ctx.ui.page("data grid with validation", {
    content: [
      ctx.ui.inputs.dataGrid("inventory", {
        data: [
          { id: "item-1", sku: "SKU001", quantity: 10, price: 99.99 },
          { id: "item-2", sku: "SKU002", quantity: 5, price: 149.99 },
          { id: "item-3", sku: "SKU003", quantity: 0, price: 199.99 },
        ],
        columns: [
          { key: "id", label: "ID", type: "id", editable: false },
          { key: "sku", label: "SKU", type: "text", editable: true },
          { key: "quantity", label: "Qty", type: "number", editable: true },
          { key: "price", label: "Price", type: "number", editable: true },
        ],
        allowAddRows: true,
        allowDeleteRows: true,
        validate: (data) => {
          // Validate that total quantity doesn't exceed 100
          const total = data.reduce((sum, item) => sum + item.quantity, 0);
          if (total > 100) {
            return "Total quantity cannot exceed 100 items";
          }

          // Validate that all quantities are non-negative
          const hasNegative = data.some((item) => item.quantity < 0);
          if (hasNegative) {
            return "Quantities must be non-negative";
          }

          // Validate that at least one item exists
          if (data.length === 0) {
            return "At least one item must be present";
          }

          // Validate that all prices are positive
          const hasInvalidPrice = data.some((item) => item.price <= 0);
          if (hasInvalidPrice) {
            return "All prices must be greater than zero";
          }

          return true;
        },
      }),
    ],
  });

  // Test 3: DataGrid with type coercion
  await ctx.ui.page("data grid with types", {
    content: [
      ctx.ui.inputs.dataGrid("orders", {
        data: [
          {
            orderId: "ORD-001",
            customerName: "John Doe",
            orderTotal: 250.5,
            isPaid: true,
          },
          {
            orderId: "ORD-002",
            customerName: "Jane Smith",
            orderTotal: 125.75,
            isPaid: false,
          },
        ],
        columns: [
          { key: "orderId", label: "Order ID", type: "text" },
          { key: "customerName", label: "Customer", type: "text" },
          { key: "orderTotal", label: "Total", type: "number" },
          { key: "isPaid", label: "Paid", type: "boolean" },
        ],
        allowAddRows: false,
        allowDeleteRows: false,
      }),
    ],
  });

  return null;
});
