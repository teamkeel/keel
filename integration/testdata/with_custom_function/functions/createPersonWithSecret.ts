import { models, CreatePersonWithSecret } from "@teamkeel/sdk";

export default CreatePersonWithSecret((ctx, inputs) => {
  return models.person.create({
    ...inputs,
    name: ctx.secrets.NAME_API_KEY,
  });
});
