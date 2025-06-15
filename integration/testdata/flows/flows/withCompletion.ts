import { WithCompletion } from "@teamkeel/sdk";


export default WithCompletion({}, async (ctx) => {
  await ctx.step("my step", async () => {
    return;
  });

  ctx.complete({
    title: "hello",
    content: [],
    data: {},
  });
});
