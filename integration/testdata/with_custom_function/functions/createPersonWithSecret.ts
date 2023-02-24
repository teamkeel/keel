import { CreatePersonWithSecret } from "@teamkeel/sdk";

export default CreatePersonWithSecret((inputs, api, ctx) => {
  return api.models.person.create({
    ...inputs,
    name: ctx.secrets.NAME_API_KEY,
  });
});
