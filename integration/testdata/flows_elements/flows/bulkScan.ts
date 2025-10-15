import { BulkScan, FlowConfig } from "@teamkeel/sdk";
import { ColumnNode } from "kysely";

const config = {
  // See https://docs.keel.so/flows for options
} as const satisfies FlowConfig;

export default BulkScan(config, async (ctx) => {
  const page1 = await ctx.ui.page("multi scan page", {
    content: [
      ctx.ui.inputs.scan("bulkScan", {
        mode: "multi",
        duplicateHandling: "rejectDuplicates",
      }),
    ],
  });

  if (page1.bulkScan.length != 3) {
    throw new Error("3 scans expected");
  }

  const page2 = await ctx.ui.page("single scan page", {
    content: [ctx.ui.inputs.scan("singleScan", { mode: "single" })],
    actions: ["finish"],
  });

  if (page2.data.singleScan !== "abc") {
    throw new Error("abc expected");
  }

  if (page2.action !== "finish") {
    throw new Error("finish action expected");
  }

  return null;
});
