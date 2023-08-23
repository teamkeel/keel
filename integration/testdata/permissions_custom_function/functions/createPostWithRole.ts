import { CreatePostWithRole } from "@teamkeel/sdk";

export default CreatePostWithRole({
  beforeWrite: async (ctx, inputs) => {
    return {
      title: inputs.title,
      businessId: inputs.business.id,
    };
  },
});
