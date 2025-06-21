import { ScheduledRunManually, models } from "@teamkeel/sdk";

export default ScheduledRunManually(async (ctx) => {
  const track = await models.trackJob.update(
    { id: "55555" },
    { didJobRun: true }
  );
  if (track == null) {
    throw new Error("expected row");
  }
});
