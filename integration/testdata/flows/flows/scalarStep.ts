import { ScalarStep, models } from "@teamkeel/sdk";

export default ScalarStep({}, async (ctx) => {
  await ctx.step("scalar step", async () => {
    return 10;
  });
});
