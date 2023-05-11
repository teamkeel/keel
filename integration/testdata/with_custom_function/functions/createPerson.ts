import { models, permissions, CreatePerson } from "@teamkeel/sdk";

export default CreatePerson((ctx, inputs) => {
  permissions.allow();

  return models.person.create(inputs);
});
