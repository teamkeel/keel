import { models, permissions, CountName } from "@teamkeel/sdk";

export default CountName(async (_, inputs) => {
  permissions.allow();

  const persons = await models.person.findMany({
    name: { equals: inputs.name },
  });

  return {
    count: persons.length,
  };
});
