import { ListOptionalFieldsWithOptionalInputs } from "@teamkeel/sdk";

export default ListOptionalFieldsWithOptionalInputs((inputs, api) => {
  return api.models.person.findMany(inputs.where!);
});
