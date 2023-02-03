import { CreatePerson, fetch} from "@teamkeel/sdk";

export default CreatePerson(async (inputs, api) => {

  const foo = await fetch("garbage url")

  return api.models.person.create(inputs);
});

