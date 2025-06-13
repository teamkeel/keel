import { DuplicateStepUiName } from "@teamkeel/sdk";

export default DuplicateStepUiName({}, async (ctx) => {
  await ctx.step("my step", async () => {
    return;
  });

  await ctx.ui.page("my step", {
    content: [],
  });
});
