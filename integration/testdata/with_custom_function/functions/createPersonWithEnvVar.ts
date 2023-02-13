import { CreatePersonWithEnvVar } from "@teamkeel/sdk";

export default CreatePersonWithEnvVar((inputs, api, ctx) => {
  return api.models.person.create({
    ...inputs,
    name: ctx.env.TEST,
  });
});
