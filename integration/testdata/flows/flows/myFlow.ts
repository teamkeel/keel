import { MyFlow, models } from "@teamkeel/sdk";

export default MyFlow(async (ctx, inputs) => {
  const thing = await ctx.step("insert thing", async () => {
    return await models.thing.create({
      name: inputs.name,
    });
  });

  const values = await ctx.ui.page({
    title: "My Flow",
    description: "This is a description",
    content: [
      ctx.ui.inputs.text("name", {
        label: "Name",
      }),
      ctx.ui.display.divider(),
      ctx.ui.inputs.number("age", {
        label: "Age",
      }),
    ],
  });

  await ctx.step("update thing", async () => {
    return await models.thing.update(
      { id: thing.id },
      {
        name: values.name,
        age: values.age,
      }
    );
  });
});
