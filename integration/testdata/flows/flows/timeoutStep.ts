import { TimeoutStep } from "@teamkeel/sdk";

export default TimeoutStep({}, async (ctx) => {
  await ctx.step("timeout step", { timeoutInMs: 1 }, async () => {
    await new Promise((resolve) => setTimeout(resolve, 100));
  });
});
