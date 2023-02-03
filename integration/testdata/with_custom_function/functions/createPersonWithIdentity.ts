import { CreatePersonWithIdentity } from "@teamkeel/sdk";

export default CreatePersonWithIdentity((inputs, api, ctx) => {
  var identityId = ctx.identity.id;
  var identityEmail = ctx.identity.email;

  return api.models.person.create({
    name: identityEmail,
    gender: inputs["gender"],
    niNumber: inputs["niNumber"],
    identityId: identityId,
  });
});
