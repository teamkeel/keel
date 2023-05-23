import {
  models,
  permissions,
  ListOptionalFieldsWithOptionalInputs,
} from "@teamkeel/sdk";

export default ListOptionalFieldsWithOptionalInputs((_, inputs) => {
  permissions.allow();

  return models.person.findMany({
    where: inputs.where!,
  });
});
