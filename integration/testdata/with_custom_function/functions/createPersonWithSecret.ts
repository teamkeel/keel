import { CreatePersonWithSecret } from "@teamkeel/sdk";

export default CreatePersonWithSecret((inputs, api, ctx) => {
  api.permissions.allow();

  return api.models.person.create({
    ...inputs,
    name: ctx.secrets.NAME_API_KEY,
  });
});
