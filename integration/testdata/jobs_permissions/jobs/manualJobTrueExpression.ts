import { ManualJobTrueExpression, models } from "@teamkeel/sdk";

export default ManualJobTrueExpression(async (ctx, inputs) => {
  const track = await models.trackJob.update(inputs, { didJobRun: true });
  if (track == null) {
    throw new Error("expected row");
  }
});
