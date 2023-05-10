import { CreatePerson } from "@teamkeel/sdk";

export default CreatePerson((ctx, inputs, api) => {
  api.permissions.allow();

  return api.models.person.create(inputs);
});
