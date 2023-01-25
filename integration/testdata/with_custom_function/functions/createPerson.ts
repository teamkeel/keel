import { CreatePerson } from "@teamkeel/sdk";

export default CreatePerson((inputs, api) => {
  return api.models.person.create(inputs);
});
