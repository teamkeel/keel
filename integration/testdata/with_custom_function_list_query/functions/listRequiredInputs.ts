import { ListRequiredInputs } from "@teamkeel/sdk";

export default ListRequiredInputs((inputs, api) => {
  api.permissions.allow();

  return api.models.person.findMany(inputs.where);
});
