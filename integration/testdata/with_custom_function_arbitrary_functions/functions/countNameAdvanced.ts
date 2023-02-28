import { CountNameAdvanced } from "@teamkeel/sdk";

export default CountNameAdvanced(async (inputs, api, ctx) => {
  var persons = await api.models.person.findMany({
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
