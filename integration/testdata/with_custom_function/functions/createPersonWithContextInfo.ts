import { CreatePersonWithContextInfo } from "@teamkeel/sdk";

export default CreatePersonWithContextInfo({
  beforeWrite: async (ctx, inputs) => {
    const { identity } = ctx;
    return {
      name: identity != null ? identity.email! : "none",
      gender: inputs.gender,
      niNumber: inputs.niNumber,
      identityId: identity != null ? identity.id : null,
      ctxNow: ctx.now(),
    };
  },
});
