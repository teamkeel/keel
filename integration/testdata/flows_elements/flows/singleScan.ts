import { SingleScan, FlowConfig } from "@teamkeel/sdk";

const config = {
  // See https://docs.keel.so/flows for options
} as const satisfies FlowConfig;

export default SingleScan(config, async (ctx) => {
  const page1 = await ctx.ui.page("single scan page", {
    content: [
      ctx.ui.inputs.scan("barcode", {
        mode: "single",
        title: "Scan Barcode",
        description: "Please scan the product barcode",
        validate: (data) => {
          // Barcode must be at least 8 characters
          if (data.length < 8) {
            return "Barcode must be at least 8 characters";
          }
          // Barcode must be alphanumeric
          if (!/^[a-zA-Z0-9]+$/.test(data)) {
            return "Barcode must be alphanumeric";
          }
          return true;
        },
      }),
    ],
  });

  return {
    barcode: page1.barcode,
  };
});
