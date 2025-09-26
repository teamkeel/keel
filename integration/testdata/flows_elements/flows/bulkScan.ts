import { BulkScan, FlowConfig } from "@teamkeel/sdk";

const config = {
  // See https://docs.keel.so/flows for options
} as const satisfies FlowConfig;

export default BulkScan(config, async (ctx) => {
  await ctx.ui.page("bulkScan page without actions", {
    content: [ctx.ui.interactive.bulkScan("bulkScan")],
  });

  await ctx.ui.page("bulkScan page with actions", {
    content: [ctx.ui.interactive.bulkScan("bulkScan")],
    actions: ["finish"],
  });

  return null;
});
