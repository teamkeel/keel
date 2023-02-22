import { ListRequiredInputs } from "@teamkeel/sdk";

export default ListRequiredInputs((inputs, api) => {
  return api.models.person.findMany(inputs.where);
});
