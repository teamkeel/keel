import { CreateDbPermissionFn } from "@teamkeel/sdk";

export default CreateDbPermissionFn({
  beforeWrite: async (ctx, inputs, values) => {
    return {
      title: inputs.title,
      lastUpdatedById: inputs.lastUpdatedBy!.id!,
    };
  },
});
