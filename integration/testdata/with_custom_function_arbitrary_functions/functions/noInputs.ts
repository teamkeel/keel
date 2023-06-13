import { models, permissions, NoInputs } from "@teamkeel/sdk";

export default NoInputs(async (_, inputs) => {
  permissions.allow();
  return { success: true };
});
