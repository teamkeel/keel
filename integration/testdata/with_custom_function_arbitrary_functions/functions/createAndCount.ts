import { models, permissions, CreateAndCount } from "@teamkeel/sdk";

export default CreateAndCount(async (_, inputs) => {
  permissions.allow();

  const person = await models.person.create({ name: inputs.name });
  const persons = await models.person.findMany({
    where: {
      name: { equals: person.name },
    },
  });

  return {
    count: persons.length,
  };
});
