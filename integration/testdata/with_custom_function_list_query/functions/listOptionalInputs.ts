import { ListOptionalInputs } from "@teamkeel/sdk";

export default ListOptionalInputs((inputs, api) => {
  api.permissions.allow();

  return api.models.person.findMany(inputs.where!);
});
