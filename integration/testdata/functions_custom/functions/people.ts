import { models, permissions, People, Person } from "@teamkeel/sdk";

export default People(async (_, inputs) => {
  permissions.allow();

  const people: Person[] = [];

  for (let i = 0; i < inputs.ids.length; i++) {
    const person = await models.person.findOne({
      id: inputs.ids[i],
    });

    people.push(person!);
  }

  return {
    people: people,
  };
});
