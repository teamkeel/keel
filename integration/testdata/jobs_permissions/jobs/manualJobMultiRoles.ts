import { ManualJobMultiRoles, models } from "@teamkeel/sdk";

export default ManualJobMultiRoles(async (ctx, inputs) => {
    const track = await models.trackJob.update(inputs, { didJobRun: true });
    if (track == null) {
        throw new Error("expected row");
    }
});