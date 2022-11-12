import { GetPerson } from "@teamkeel/sdk";

export default GetPerson(async (inputs, api) => {
  const { object, errors } = await api.models.person.findOne(inputs);
  return { object, errors };
});
