import { models, permissions, NoInputs } from "@teamkeel/sdk";

export default NoInputs(async (ctx, inputs) => {
  permissions.allow();
  return { success: true };
});
