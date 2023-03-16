import { CreatePerson } from "@teamkeel/sdk";

export default CreatePerson((inputs, api, ctx) => {
  api.permissions.allow();
  return api.models.person.create(inputs);
});
