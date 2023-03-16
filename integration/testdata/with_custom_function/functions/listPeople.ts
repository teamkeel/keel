import { ListPeople } from "@teamkeel/sdk";

export default ListPeople((inputs, api, ctx) => {
  api.permissions.allow();

  return api.models.person.findMany(inputs.where);
});
