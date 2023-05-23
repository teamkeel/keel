import {
  models,
  permissions,
  ListOptionalFieldsWithRequiredInputs,
} from "@teamkeel/sdk";

export default ListOptionalFieldsWithRequiredInputs((_, inputs) => {
  permissions.allow();

  return models.person.findMany(inputs!);
});
