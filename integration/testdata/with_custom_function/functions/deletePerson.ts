import { DeletePerson } from "@teamkeel/sdk";

export default DeletePerson(async (inputs, api) => {
  return await api.models.person.delete(inputs.id);
});
