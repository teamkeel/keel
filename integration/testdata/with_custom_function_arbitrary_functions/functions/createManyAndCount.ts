import { CreateManyAndCount } from "@teamkeel/sdk";

export default CreateManyAndCount(async (inputs, api, ctx) => {
  let count = 0;

  for (let i = 0; i < inputs.names.length; i++) {
    const person = await api.models.person.create({ name: inputs.names[i] });
    const persons = await api.models.person.findMany({
      name: { equals: person.name },
    });
    count += persons.length;
  }

  return {
    count: count,
  };
});
