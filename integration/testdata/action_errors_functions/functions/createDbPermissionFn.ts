import { CreateDbPermissionFn } from "@teamkeel/sdk";

export default CreateDbPermissionFn({
  beforeWrite: async (ctx, inputs) => {
    return {
      title: inputs.title,
      lastUpdatedById: inputs.lastUpdatedBy!.id!,
    };
  },
});
