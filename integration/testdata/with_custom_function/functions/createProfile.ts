import { CreateProfile } from "@teamkeel/sdk";

export default CreateProfile((inputs, api) => {
  return api.models.profile.create(inputs);
});
