import { models, CreatePersonWithContextInfo } from "@teamkeel/sdk";

export default CreatePersonWithContextInfo((ctx, inputs) => {
  const { identity } = ctx;

  return models.person.create({
    name: identity != null ? identity.email! : "none",
    gender: inputs.gender,
    niNumber: inputs.niNumber,
    identityId: identity != null ? identity.id : null,
    ctxNow: ctx.now(),
  });
});
