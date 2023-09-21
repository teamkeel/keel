import { CreatePermittedFn } from "@teamkeel/sdk";

export default CreatePermittedFn({
  beforeWrite: async (ctx, inputs) => {
    return {
      title: inputs.title,
      lastUpdatedById: inputs.lastUpdatedBy!.id!,
    };
  },
});
