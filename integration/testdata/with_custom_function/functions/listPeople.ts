import { ListPeople } from "@teamkeel/sdk";

export default ListPeople((ctx, inputs, api) => {
  api.permissions.allow();

  return api.models.person.findMany(inputs.where);
});
