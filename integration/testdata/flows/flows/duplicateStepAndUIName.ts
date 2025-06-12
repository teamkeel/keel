import { DuplicateStepAndUiName } from "@teamkeel/sdk";

export default DuplicateStepAndUiName({}, async (ctx) => {
  await ctx.step("my step", async () => {
    return;
  });

  await ctx.ui.page("my step", {
    content: [],
  });
});
