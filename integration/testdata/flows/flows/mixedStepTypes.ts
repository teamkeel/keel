import { MixedStepTypes, models } from "@teamkeel/sdk";

export default MixedStepTypes(
  {
    title: "Mixed Step Types",
    description: "This is a description",
    stages: [
      {
        key: "stage1",
        name: "My stage 1",
        description: "This is stage 1's description",
      },
      {
        key: "stage2",
        name: "My stage 2",
        description: "This is stage 2's description",
      },
    ],
  },
  async (ctx, inputs) => {
    const thing = await ctx.step(
      "insert thing",
      { stage: "stage1" },
      async () => {
        const thing = await models.thing.create({
          name: inputs.name,
          age: inputs.age,
        });

        return { id: thing.id };
      }
    );

    const values = await ctx.ui.page("confirm thing", {
      title: "Update thing",
      stage: "stage1",
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

    await ctx.step("update thing", { stage: "stage2" }, async () => {
      const updated = await models.thing.update(
        { id: thing.id },
        {
          name: values.name,
          age: values.age,
        }
      );

      return { name: updated.name, age: updated.age };
    });
  }
);
