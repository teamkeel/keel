import { UpdatePerson } from "@teamkeel/sdk";

export default UpdatePerson((inputs, api, ctx) => {
  api.permissions.allow();

  return api.models.person.update(inputs.where, inputs.values);
});
