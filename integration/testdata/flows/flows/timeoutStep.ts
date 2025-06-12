import { TimeoutStep } from "@teamkeel/sdk";

export default TimeoutStep({}, async (ctx) => {
  await ctx.step("timeout step", { timeout: 10 }, async () => {
    await new Promise((resolve) => setTimeout(resolve, 100));
  });
});
