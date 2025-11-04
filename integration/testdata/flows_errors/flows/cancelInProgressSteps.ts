import { CancelInProgressSteps } from "@teamkeel/sdk";

export default CancelInProgressSteps({}, async (ctx) => {
  // First step will complete successfully
  await ctx.step("step 1", { retries: 0 }, async () => {
    return "step1 complete";
  });

  // Second step will fail after exhausting retries
  await ctx.step("step 2", { retries: 2 }, async () => {
    throw new Error("step 2 failed");
  });

  // This third step should never execute and should remain as NEW
  // When the flow fails, this step should be cancelled
  await ctx.step("step 3", { retries: 0 }, async () => {
    return "step3 complete";
  });
});
