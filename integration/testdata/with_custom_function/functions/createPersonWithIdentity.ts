import { CreatePersonWithIdentity } from "@teamkeel/sdk";

export default CreatePersonWithIdentity((inputs, api, ctx) => {
  const { email, id } = ctx.identity;

  return api.models.person.create({
    name: email,
    gender: inputs.gender,
    niNumber: inputs.niNumber,
    identityId: id,
  });
});
