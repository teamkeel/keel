import { TimeoutStep } from "@teamkeel/sdk";

export default TimeoutStep({}, async (ctx) => {
  await ctx.step("timeout step", { timeout: 10 }, async (args) => {
    if (args.stepOptions.timeout !== 10) {
      throw new Error("Should have 10ms timeout");
    }

    // Wait for 100ms (longer than the max timeout of 10ms)
    await new Promise((resolve) => setTimeout(resolve, 100));
  });
});
