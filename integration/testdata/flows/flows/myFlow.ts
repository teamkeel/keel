import { MyFlow, models } from "@teamkeel/sdk";

export default MyFlow(async (ctx, inputs) => {
  const thing = await ctx.step("insert thing", async () => {
    return await models.thing.create({
      name: inputs.name,
    });
  });

  const age = await ctx.step("update thing", async () => {
    return await models.thing.update(
      { id: thing.id },
      {
        age: inputs.age,
      }
    );
  });
});
