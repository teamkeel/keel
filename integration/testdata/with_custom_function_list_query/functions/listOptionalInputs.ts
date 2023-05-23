import { models, ListOptionalInputs } from "@teamkeel/sdk";

export default ListOptionalInputs((_, inputs) => {
  return models.person.findMany({
    where: inputs.where!,
  });
});
