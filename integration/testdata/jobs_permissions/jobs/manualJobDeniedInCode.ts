import { ManualJobDeniedInCode, models, permissions } from "@teamkeel/sdk";

export default ManualJobDeniedInCode(async (ctx, inputs) => {
  await models.trackJob.update({ id: inputs.id }, { didJobRun: true });

  if (inputs.denyIt) {
    permissions.deny();
  } else {
    permissions.allow();
  }
});
