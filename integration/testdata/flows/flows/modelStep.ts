import { ModelStep, models } from "@teamkeel/sdk";

export default ModelStep(
  {
    title: "Model step",
  },
  async (ctx, inputs) => {
    const result = await ctx.step("create and return model", async () => {
      const thing = await models.thing.create({
        name: inputs.name,
        age: inputs.age,
      });

      // Return the model instance directly (cast to any since Thing doesn't satisfy JsonSerializable constraint)
      return thing as any;
    });

    // Also test that we can use the returned model and that Date fields are properly deserialized
    await ctx.step("verify model", async () => {
      // Cast to any since we know the runtime type will have these properties
      const model = result as any;
      
      // The returned value should have the model properties
      return {
        hasId: !!model.id,
        hasName: model.name === inputs.name,
        hasAge: model.age === inputs.age,
        createdAtIsDate: model.createdAt instanceof Date,
        updatedAtIsDate: model.updatedAt instanceof Date,
        canCallDateMethod: typeof model.createdAt.getTime === 'function',
      };
    });
  }
);

