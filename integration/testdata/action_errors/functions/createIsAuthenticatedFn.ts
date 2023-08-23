import { CreateIsAuthenticatedFn } from "@teamkeel/sdk";

export default CreateIsAuthenticatedFn({
  beforeWrite: async (ctx, inputs) => {
    return {
      title: inputs.title,
      lastUpdatedById: ctx.identity!.id,
    };
  },
});
