import { CreatePostWithRole } from "@teamkeel/sdk";

export default CreatePostWithRole({
  beforeWrite: async (ctx, inputs, values) => {
    return {
      title: inputs.title,
      businessId: inputs.business.id,
    };
  },
});
