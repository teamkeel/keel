import { CreateProfile } from "@teamkeel/sdk";

export default CreateProfile((inputs, api) => {
  api.permissions.allow();

  return api.models.profile.create(inputs);
});
