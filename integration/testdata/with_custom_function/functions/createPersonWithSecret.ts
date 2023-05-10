import { CreatePersonWithSecret } from "@teamkeel/sdk";

export default CreatePersonWithSecret((ctx, inputs, api) => {
  api.permissions.allow();

  return api.models.person.create({
    ...inputs,
    name: ctx.secrets.NAME_API_KEY,
  });
});
