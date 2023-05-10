import { ListOptionalInputs } from "@teamkeel/sdk";

export default ListOptionalInputs((_, inputs, api) => {
  api.permissions.allow();

  return api.models.person.findMany(inputs.where!);
});
