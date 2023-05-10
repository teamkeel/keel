import { CreateProfile } from "@teamkeel/sdk";

export default CreateProfile((ctx, inputs, api) => {
  api.permissions.allow();

  return api.models.profile.create({ personId: inputs.person.id });
});
