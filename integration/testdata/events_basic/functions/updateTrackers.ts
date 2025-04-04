import { models, UpdateTrackers } from "@teamkeel/sdk";

export default UpdateTrackers(async (ctx, inputs) => {
  const trackers = await models.tracker.findMany();
  for (const t of trackers) {
    await models.tracker.update({ id: t.id }, { views: t.views + 1 });
  }
});
