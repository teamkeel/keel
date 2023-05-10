import { CreatePersonWithEnvVar } from "@teamkeel/sdk";

export default CreatePersonWithEnvVar((ctx, inputs, api) => {
  api.permissions.allow();

  return api.models.person.create({
    ...inputs,
    name: ctx.env.TEST,
  });
});
