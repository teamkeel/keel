import { CreateFromFile, permissions, models } from "@teamkeel/sdk";

// To learn more about what you can do with custom functions, visit https://docs.keel.so/functions
export default CreateFromFile(async (ctx, inputs) => {
  permissions.allow();
  const person = await models.person.create({
    photo: inputs.file,
    name: "Pedro",
    height: 100,
  });

  return { person };
});
