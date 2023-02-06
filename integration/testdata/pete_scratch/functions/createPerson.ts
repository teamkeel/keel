import { CreatePerson } from "@teamkeel/sdk";

export default CreatePerson((inputs, api) => {
  const foo = api.petesFunction(42)
  return api.models.person.create(inputs);
});
