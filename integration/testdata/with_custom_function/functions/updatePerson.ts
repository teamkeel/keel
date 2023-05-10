import { UpdatePerson } from "@teamkeel/sdk";

export default UpdatePerson((ctx, inputs, api) => {
  api.permissions.allow();

  return api.models.person.update(inputs.where, inputs.values);
});
