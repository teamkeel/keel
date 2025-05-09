import { ErrorInStep } from "@teamkeel/sdk";

export default ErrorInStep({}, async (ctx) => {
  await ctx.step("erroring step", { maxRetries: 3 }, async () => {
    throw new Error("Error in step");
  });
});
