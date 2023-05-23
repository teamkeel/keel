import { models, permissions, CountNameAdvanced } from "@teamkeel/sdk";

export default CountNameAdvanced(async (_, inputs) => {
  permissions.allow();

  const persons = await models.person.findMany({
    where: {
      name: {
        startsWith: inputs.startsWith,
        contains: inputs.contains,
        endsWith: inputs.endsWith,
      },
    },
  });

  return {
    count: persons.length,
  };
});
