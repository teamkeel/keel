import { WithReturnedData } from "@teamkeel/sdk";

export default WithReturnedData({}, async (ctx) => {
  await ctx.step("my step", async () => {
    return;
  });

  return "hello";
});
