import { CountName } from "@teamkeel/sdk";

export default CountName(async (inputs, api, ctx) => {
  api.permissions.allow();

  const persons = await api.models.person.findMany({
    name: { equals: inputs.name },
  });

  return {
    count: persons.length,
  };
});
