import { ListOptionalFieldsWithOptionalInputs } from "@teamkeel/sdk";

export default ListOptionalFieldsWithOptionalInputs((_, inputs, api) => {
  api.permissions.allow();

  return api.models.person.findMany(inputs.where!);
});
