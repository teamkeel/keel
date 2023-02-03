import { ListPeople } from "@teamkeel/sdk";

export default ListPeople((inputs, api) => {
  return api.models.person.findMany(inputs.where);
});
