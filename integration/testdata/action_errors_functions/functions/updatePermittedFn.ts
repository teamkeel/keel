import { UpdatePermittedFn } from "@teamkeel/sdk";

export default UpdatePermittedFn({
  beforeWrite: async (ctx, inputs) => {
    return {
      title: inputs.values.title,
      lastUpdatedById: inputs.values.lastUpdatedBy?.id,
    };
  },
});
