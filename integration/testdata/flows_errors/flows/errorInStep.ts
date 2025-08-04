import { ErrorInStep } from "@teamkeel/sdk";

export default ErrorInStep({}, async (ctx) => {
  await ctx.step(
    "erroring step",
    { retries: 2, retryDelay: 3000 },
    async (args) => {
      if (args.stepOptions.retries !== 2) {
        throw new Error("Should have 2 retries");
      }

      throw new Error("Error in step");
    }
  );
});
