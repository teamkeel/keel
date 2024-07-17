import { CreatePerson, permissions, models } from "@teamkeel/sdk";

// To learn more about what you can do with custom functions, visit https://docs.keel.so/functions
export default CreatePerson(async (_, inputs) => {
  permissions.allow();

  const response = await models.person.create({
    name: inputs.name,
    height: inputs.height,
  });

  if (response.height) {
    return {
      id: response.id,
      name: response.name,
      height: response.height,
    };
  } else {
    return {
      id: response.id,
      name: response.name,
    };
  }
});
