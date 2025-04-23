import { MyFlow, models } from "@teamkeel/sdk";

export default MyFlow({
  title: "My Flow",
  description: "This is a description",
  stages: [
    {
      key: "stage 1",
      name: "My stage 1",
      description: "This is a descriptio 1",
    },
    {
      key: "stage 2",
      name: "My stage 2",
      description: "This is a description 2",
    },
  ],
}, async (ctx, inputs) => {
  const thingId = await ctx.step("insert thing", async () => {
    const thing = await models.thing.create({
      name: inputs.name,
      age: inputs.age,
    });

    return thing.id
  });

  const values = await ctx.ui.page({
    title: "Update thing",
    stage: "Update",
    description: "Overwrite the existing data in thing",
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

  console.log("values", values);

  await ctx.step("update thing", async () => {
    return await models.thing.update(
      { id: thingId },
      {
        name: values.name,
        age: values.age,
      }
    );
  });
});
