import { models, ListRequiredInputs } from "@teamkeel/sdk";

export default ListRequiredInputs((_, inputs) => {
  return models.person.findMany(inputs);
});
