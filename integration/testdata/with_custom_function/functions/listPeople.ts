import { ListPeople } from "@teamkeel/sdk";

export default ListPeople(async (inputs, api) => {
  return await api.models.person.findMany(inputs);
});
