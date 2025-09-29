import { BulkScan, FlowConfig } from "@teamkeel/sdk";

const config = {
  // See https://docs.keel.so/flows for options
} as const satisfies FlowConfig;

export default BulkScan(config, async (ctx) => {
  await ctx.ui.page("multi scan page", {
    content: [
      ctx.ui.inputs.scan("bulkScan", {
        mode: "multi",
        duplicateHandling: "rejectDuplicates",
      }),
    ],
  });

  await ctx.ui.page("single scan page", {
    content: [ctx.ui.inputs.scan("singleScan", { mode: "single" })],
    actions: ["finish"],
  });

  return null;
});
