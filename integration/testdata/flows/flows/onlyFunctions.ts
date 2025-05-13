import { OnlyFunctions, models } from "@teamkeel/sdk";

export default OnlyFunctions(
  {
    title: "Flow with two functions",
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
    const thingId = await ctx.step(
      "insert thing",
      {
        stage: "stage1",
      },
      async () => {
        const thing = await models.thing.create({
          name: inputs.name,
          age: inputs.age,
        });

        return thing.id;
      }
    );

    await ctx.step(
      "update thing",
      {
        stage: "stage2",
      },
      async () => {
        const thing = await models.thing.findOne({
          id: thingId,
        });

        const updated = await models.thing.update(
          { id: thing!.id },
          {
            name: thing!.name + " Updated",
            age: thing!.age! + 1,
          }
        );

        return { name: updated.name, age: updated.age };
      }
    );
  }
);
