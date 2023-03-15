import { ListOptionalFieldsWithOptionalInputs } from "@teamkeel/sdk";

export default ListOptionalFieldsWithOptionalInputs((inputs, api) => {
  api.permissions.allow();

  return api.models.person.findMany(inputs.where!);
});
