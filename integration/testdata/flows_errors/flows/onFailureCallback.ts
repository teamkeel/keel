import { OnFailureCallback, models } from "@teamkeel/sdk";

export default OnFailureCallback({}, async (ctx) => {
  await ctx.step(
    "erroring step",
    {
      retries: 2,
      onFailure: async () => {
        const things = await models.thing.findMany();

        if (things.length != 3) {
          throw new Error("There should be 3 things");
        }

        for (const thing of things) {
          await models.thing.delete({ id: thing.id });
        }
      },
    },
    async (a) => {
      await models.thing.create({
        name: "test",
      });
      const things = await models.thing.findMany();

      throw new Error(things.length.toString() + " exists");
    }
  );
});
