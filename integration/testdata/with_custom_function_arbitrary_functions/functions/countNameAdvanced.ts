import { CountNameAdvanced } from "@teamkeel/sdk";

export default CountNameAdvanced(async (_, inputs, api) => {
  api.permissions.allow();

  const persons = await api.models.person.findMany({
    name: {
      startsWith: inputs.startsWith,
      contains: inputs.contains,
      endsWith: inputs.endsWith,
    },
  });

  return {
    count: persons.length,
  };
});
