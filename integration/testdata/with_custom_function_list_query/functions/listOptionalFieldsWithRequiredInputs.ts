import { ListOptionalFieldsWithRequiredInputs } from "@teamkeel/sdk";

export default ListOptionalFieldsWithRequiredInputs((_, inputs, api) => {
  api.permissions.allow();

  return api.models.person.findMany(inputs.where!);
});
