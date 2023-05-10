import { CreateAndCount } from "@teamkeel/sdk";

export default CreateAndCount(async (_, inputs, api) => {
  api.permissions.allow();

  const person = await api.models.person.create({ name: inputs.name });
  const persons = await api.models.person.findMany({
    name: { equals: person.name },
  });

  return {
    count: persons.length,
  };
});
