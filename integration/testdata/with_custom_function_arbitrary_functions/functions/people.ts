import { People, Person } from "@teamkeel/sdk";

export default People(async (inputs, api, ctx) => {
  api.permissions.allow();

  const people: Person[] = [];

  for (let i = 0; i < inputs.ids.length; i++) {
    const person = await api.models.person.findOne({
      id: inputs.ids[i],
    });

    people.push(person!);
  }

  return {
    people: people,
  };
});
