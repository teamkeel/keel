import { models, CreateProfile } from "@teamkeel/sdk";

export default CreateProfile((ctx, inputs) => {
  return models.profile.create({ personId: inputs.person.id });
});
