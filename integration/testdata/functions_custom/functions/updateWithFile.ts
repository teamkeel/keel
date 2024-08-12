import { UpdateWithFile, permissions, models } from "@teamkeel/sdk";

// To learn more about what you can do with custom functions, visit https://docs.keel.so/functions
export default UpdateWithFile(async (ctx, inputs) => {
  permissions.allow();
  const person = await models.person.create({
    name: "Pedro",
    height: 100,
  });

  const updatedPerson = await models.person.update(
    { id: person.id },
    {
      photo: inputs.file,
    }
  );

  return { person: updatedPerson };
});
