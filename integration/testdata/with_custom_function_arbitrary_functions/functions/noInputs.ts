import { models, permissions, RandomName as NoInputs } from "@teamkeel/sdk";

export default NoInputs(async (_, inputs) => {
  permissions.allow();

  return;
});
