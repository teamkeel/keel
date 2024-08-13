import { ReadFile, models, permissions } from "@teamkeel/sdk";

// To learn more about what you can do with custom functions, visit https://docs.keel.so/functions
export default ReadFile(async (ctx, inputs) => {
  permissions.allow();
  const person = await models.person.findOne({ id: inputs.id });
  const contents = await person?.photo?.read();

  return {
    photo: person?.photo,
    contents: contents?.toString("base64"),
  };
});
