import { MixedStepTypes, models } from "@teamkeel/sdk";

export default MixedStepTypes({}, async (ctx, inputs) => {
  const thing = await ctx.step("insert thing", async () => {
    const thing = await models.thing.create({
      name: inputs.name,
      age: inputs.age,
    });

    return { id: thing.id };
  });

  const values = await ctx.ui.page("confirm thing", {
    title: "Update thing",
    description: "Confirm the existing data in thing",
    content: [
      ctx.ui.inputs.text("name", {
        label: "Name",
        defaultValue: inputs.name,
      }),
      ctx.ui.display.divider(),
      ctx.ui.inputs.number("age", {
        label: "Age",
        defaultValue: inputs.age,
      }),
    ],
  });

  await ctx.step("update thing", async () => {
    const updated = await models.thing.update(
      { id: thing.id },
      {
        name: values.name,
        age: values.age,
      }
    );

    return { name: updated.name, age: updated.age };
  });
});
