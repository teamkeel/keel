import { CreatePerson } from "@teamkeel/sdk";

export default CreatePerson(async (inputs, api) => {
  return await api.models.person.create(inputs);
});
