import { CreatePersonWithContextInfo } from "@teamkeel/sdk";

export default CreatePersonWithContextInfo((inputs, api, ctx) => {
  api.permissions.allow();

  const { identity } = ctx;

  return api.models.person.create({
    name: identity != null ? identity.email! : "none",
    gender: inputs.gender,
    niNumber: inputs.niNumber,
    identityId: identity != null ? identity.id : null,
    ctxNow: ctx.now(),
  });
});
