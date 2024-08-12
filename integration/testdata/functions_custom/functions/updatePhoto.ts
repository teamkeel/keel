import { UpdatePhoto, models, permissions } from "@teamkeel/sdk";

// To learn more about what you can do with custom functions, visit https://docs.keel.so/functions
export default UpdatePhoto(async (ctx, inputs) => {
  permissions.allow();

  const updated = await models.person.update(
    { id: inputs.id },
    { photo: inputs.photo }
  );
  return { person: updated };
});
