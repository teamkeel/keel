import { ErrorInStep, RetryBackoffLinear } from "@teamkeel/sdk";

export default ErrorInStep({}, async (ctx) => {
  await ctx.step(
    "erroring step",
    { retries: 3, retryPolicy: RetryBackoffLinear(1) },
    async (args) => {
      if (args.stepOptions.retries !== 3) {
        throw new Error("Should have 3 retries");
      }

      throw new Error("Error in step");
    }
  );
});
