import { models, CreatePersonWithEnvVar } from "@teamkeel/sdk";

export default CreatePersonWithEnvVar((ctx, inputs) => {
  return models.person.create({
    ...inputs,
    name: ctx.env.TEST,
  });
});
