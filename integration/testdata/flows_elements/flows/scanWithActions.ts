import { ScanWithActions, FlowConfig } from "@teamkeel/sdk";

const config = {
  // See https://docs.keel.so/flows for options
} as const satisfies FlowConfig;

export default ScanWithActions(config, async (ctx) => {
  // Test single scan with actions
  const page1 = await ctx.ui.page("single scan with validation", {
    content: [
      ctx.ui.inputs.scan("productCode", {
        mode: "single",
        validate: (data, action) => {
          // Verify action parameter is passed correctly when an action is provided
          if (
            action !== undefined &&
            action !== "verify" &&
            action !== "lookup" &&
            action !== "skip"
          ) {
            throw new Error(
              `Expected action to be 'verify', 'lookup', or 'skip', got: ${action}`
            );
          }

          if (action === "verify") {
            // Verification requires specific prefix
            if (!data.startsWith("PROD-")) {
              return "Product code must start with 'PROD-' for verification";
            }
            // Must be at least 10 characters
            if (data.length < 10) {
              return "Product code must be at least 10 characters for verification";
            }
          }

          if (action === "lookup") {
            // Lookup is more lenient, just needs to not be empty
            if (!data || data.trim() === "") {
              return "Product code cannot be empty for lookup";
            }
          }

          return true;
        },
      }),
    ],
    actions: ["verify", "lookup", "skip"],
  });

  // Test multi scan with actions
  const page2 = await ctx.ui.page("multi scan with validation", {
    content: [
      ctx.ui.inputs.scan("barcodes", {
        mode: "multi",
        duplicateHandling: "rejectDuplicates",
        min: 1,
        max: 10,
        validate: (data, action) => {
          // Verify action parameter is passed correctly when an action is provided
          if (
            action !== undefined &&
            action !== "process" &&
            action !== "save"
          ) {
            throw new Error(
              `Expected action to be 'process' or 'save', got: ${action}`
            );
          }

          const barcodes = data as unknown as string[];

          if (action === "process") {
            // Processing requires at least 3 items
            if (barcodes.length < 3) {
              return "Must scan at least 3 items to process";
            }

            // All codes must be numeric for processing
            const allNumeric = barcodes.every((code) => /^\d+$/.test(code));
            if (!allNumeric) {
              return "All barcodes must be numeric for processing";
            }
          }

          if (action === "save") {
            // Save requires at least 1 item
            if (barcodes.length === 0) {
              return "Must scan at least 1 item to save";
            }
          }

          return true;
        },
      }),
    ],
    actions: ["process", "save"],
  });

  // Test scan with quantity tracking
  const page3 = await ctx.ui.page("scan with quantity", {
    content: [
      ctx.ui.inputs.scan("items", {
        mode: "multi",
        duplicateHandling: "trackQuantity",
        validate: (data, action) => {
          // Verify action parameter is passed correctly when an action is provided
          if (
            action !== undefined &&
            action !== "checkout" &&
            action !== "continue"
          ) {
            throw new Error(
              `Expected action to be 'checkout' or 'continue', got: ${action}`
            );
          }

          const items = data as unknown as {
            value: string;
            quantity: number;
          }[];

          if (action === "checkout") {
            // Checkout requires total quantity >= 5
            const totalQty = items.reduce(
              (sum, item) => sum + item.quantity,
              0
            );
            if (totalQty < 5) {
              return "Total quantity must be at least 5 for checkout";
            }

            // Each item must have quantity >= 2 for checkout
            const allValidQty = items.every((item) => item.quantity >= 2);
            if (!allValidQty) {
              return "Each item must have quantity of at least 2 for checkout";
            }
          }

          if (action === "continue") {
            // Continue just requires at least one scan
            if (items.length === 0) {
              return "Must scan at least 1 item";
            }
          }

          return true;
        },
      }),
    ],
    actions: ["checkout", "continue"],
  });

  return {
    action1: page1.action,
    productCode: page1.data.productCode,
    action2: page2.action,
    barcodes: page2.data.barcodes,
    action3: page3.action,
    items: page3.data.items,
  };
});
