import { ManualJobEnvExpression2, models } from "@teamkeel/sdk";

export default ManualJobEnvExpression2(async (ctx, inputs) => {
  const track = await models.trackJob.update(inputs, { didJobRun: true });
  if (track == null) {
    throw new Error("expected row");
  }
});
