import { ScheduledWithoutPermissions, models } from "@teamkeel/sdk";

export default ScheduledWithoutPermissions(async (ctx) => {
  const track = await models.trackJob.update(
    { id: "12345" },
    { didJobRun: true }
  );
  if (track == null) {
    throw new Error("expected row");
  }
});
