import { CountName } from "@teamkeel/sdk";

export default CountName(async (_, inputs, api) => {
  api.permissions.allow();

  const persons = await api.models.person.findMany({
    name: { equals: inputs.name },
  });

  return {
    count: persons.length,
  };
});
