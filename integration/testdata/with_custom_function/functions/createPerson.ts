import { CreatePerson } from "@teamkeel/sdk";

export default CreatePerson(async (inputs, api, ctx) => {
  api.permissions.allow();

  return await api.models.person.create(inputs);
});
