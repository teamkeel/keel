import { WriteCustomFunction, models } from "@teamkeel/sdk";

// To learn more about what you can do with custom functions, visit https://docs.keel.so/functions
export default WriteCustomFunction(async (ctx, inputs) => {
  const mod = await models.myDuration.create({ dur: inputs.dur });
  return {
    model: mod,
  };
});
