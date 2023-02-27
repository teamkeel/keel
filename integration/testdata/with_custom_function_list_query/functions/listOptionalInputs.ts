import { ListOptionalInputs } from "@teamkeel/sdk";

export default ListOptionalInputs((inputs, api) => {
  return api.models.person.findMany(inputs.where!);
});
