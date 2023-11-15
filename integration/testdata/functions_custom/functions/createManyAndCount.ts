import { models, permissions, CreateManyAndCount } from "@teamkeel/sdk";

export default CreateManyAndCount(async (_, inputs) => {
  permissions.allow();

  let count = 0;

  for (let i = 0; i < inputs.names.length; i++) {
    const person = await models.person.create({ name: inputs.names[i] });
    const persons = await models.person.findMany({
      where: {
        name: { equals: person.name },
      },
    });
    count += persons.length;
  }

  return {
    count: count,
  };
});
