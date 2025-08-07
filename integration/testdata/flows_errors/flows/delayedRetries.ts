import { DelayedRetries, RetryConstant } from "@teamkeel/sdk";

export default DelayedRetries({}, async (ctx) => {
  await ctx.step(
    "constant delay step",
    { retries: 2, retryPolicy: RetryConstant(2) },
    async (args) => {
      if (args.attempt !== 3) {
        throw new Error("enforce 2 retries");
      }

      return "completed";
    }
  );
});
