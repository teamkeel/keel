import { models, ListPeople } from "@teamkeel/sdk";

export default ListPeople((ctx, inputs) => {
  return models.person.findMany(inputs.where);
});
