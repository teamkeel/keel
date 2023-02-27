import { ListOptionalFieldsWithRequiredInputs } from "@teamkeel/sdk";

export default ListOptionalFieldsWithRequiredInputs((inputs, api) => {
  return api.models.person.findMany(inputs.where!);
});
