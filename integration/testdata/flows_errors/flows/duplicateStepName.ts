import { DuplicateStepName } from "@teamkeel/sdk";

export default DuplicateStepName({}, async (ctx) => {
  await ctx.step("my step", async () => {
    return;
  });

  await ctx.step("my step", async () => {
    return;
  });
});
