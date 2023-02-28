import { CountName } from "@teamkeel/sdk";

export default CountName(async (inputs, api, ctx) => {
  var persons = await api.models.person.findMany({
    name: { equals: inputs.name },
  });

  return {
    count: persons.length,
  };
});
