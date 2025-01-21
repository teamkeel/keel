import { WriteAndDuplicate, models, Duration } from "@teamkeel/sdk";

// To learn more about what you can do with custom functions, visit https://docs.keel.so/functions
export default WriteAndDuplicate(async (ctx, inputs) => {
  const mod = await models.myDuration.create({ dur: inputs.dur });
  const dbModel = await models.myDuration.findOne({ id: mod.id });

  const duplicate = await models.myDuration.create({
    dur: dbModel?.dur,
  });
  return {
    model: mod,
    duplicate: duplicate,
  };
});
