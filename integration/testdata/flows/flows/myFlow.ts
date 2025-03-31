import { MyFlow, models } from "@teamkeel/sdk";

export default MyFlow(async (ctx, inputs) => {
  await models.thing.create({
    name: inputs.name
  });
});
