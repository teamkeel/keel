import { DoNotRetry, models, NonRetriableError } from "@teamkeel/sdk";

export default DoNotRetry({}, async (ctx) => {
  await ctx.step(
    "erroring step",
    {
      retries: 2,
      onFailure: async () => {
        const things = await models.thing.findMany();

        if (things.length != 1) {
          throw new Error("There should be 1 thing");
        }

        for (const thing of things) {
          await models.thing.delete({ id: thing.id });
        }
      },
    },
    async () => {
      await models.thing.create({
        name: "test",
      });
      const things = await models.thing.findMany();

      throw new NonRetriableError("do not retry!");
    }
  );
});
