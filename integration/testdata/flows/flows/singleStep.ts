import { SingleStep, models } from "@teamkeel/sdk";

export default SingleStep({}, async (ctx) => {
  await ctx.step("insert thing", async (args) => {
    if (args.attempt !== 0) {
      throw new Error("Should be attempt 0");
    }
    if (args.stepOptions.retries !== 4) {
      throw new Error("Should have 4 retries");
    }
    if (args.stepOptions.timeout !== 60000) {
      throw new Error("Should have 60000ms timeout");
    }

    return { number: 10 };
  });
});
