import { models, CreatePost } from "@teamkeel/sdk";

export default CreatePost(async (_, inputs) => {
  const result = await models.post.create({
    title: inputs.title,
    businessId: inputs.business.id,
  });

  return result;
});
