import { CreatePost } from "@teamkeel/sdk";

export default CreatePost({
  beforeWrite: async (ctx, inputs) => {
    return {
      title: inputs.title,
      businessId: inputs.business.id,
    };
  },
});
