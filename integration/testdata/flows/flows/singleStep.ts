import { SingleStep, models } from "@teamkeel/sdk";

export default SingleStep({}, async (ctx) => {
  await ctx.step("insert thing", async () => {
    return { number: 10 };
  });
});
