import { CreatePost } from "@teamkeel/sdk";

export default CreatePost({
  beforeWrite: async (ctx, inputs, values) => {
    return {
      title: inputs.title,
      businessId: inputs.business.id,
    };
  },
});
