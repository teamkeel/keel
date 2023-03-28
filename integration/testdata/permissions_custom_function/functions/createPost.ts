import { CreatePost } from "@teamkeel/sdk";

export default CreatePost(async (inputs, api, ctx) => {
  const result = await api.models.post.create({
    title: inputs.title,
    businessId: inputs.business.id,
  });

  return result;
});
