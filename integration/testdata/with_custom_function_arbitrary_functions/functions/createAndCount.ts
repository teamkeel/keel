import { CreateAndCount } from "@teamkeel/sdk";

export default CreateAndCount(async (inputs, api, ctx) => {
  var person = await api.models.person.create({ name: inputs.name });
  var persons = await api.models.person.findMany({
    name: { equals: person.name },
  });

  return {
    count: persons.length,
  };
});
