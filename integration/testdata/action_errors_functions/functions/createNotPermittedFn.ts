import { CreateNotPermittedFn } from "@teamkeel/sdk";

export default CreateNotPermittedFn({
  beforeWrite: async (ctx, inputs) => {
    return {
      title: inputs.title,
      lastUpdatedById: ctx.identity?.id || "",
    };
  },
});
