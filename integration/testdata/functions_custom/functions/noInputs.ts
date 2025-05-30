import { models, permissions, NoInputs } from "@teamkeel/sdk";

export default NoInputs(async (ctx) => {
  permissions.allow();
  return { success: true };
});
