import { CreateManyAndCount } from "@teamkeel/sdk";

export default CreateManyAndCount(async (inputs, api, ctx) => {
  var count = 0;

  for (let i = 0; i < inputs.names.length; i++) {
    var person = await api.models.person.create({ name: inputs.names[i] });
    var persons = await api.models.person.findMany({
      name: { equals: person.name },
    });
    count += persons.length;
  }

  return {
    count: count,
  };
});
