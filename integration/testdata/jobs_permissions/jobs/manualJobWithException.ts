import { ManualJob, models } from "@teamkeel/sdk";

export default ManualJob(async (ctx, inputs) => {
  await models.trackJob.update(inputs, { didJobRun: true });
  throw new Error("something bad has happened!");
});
