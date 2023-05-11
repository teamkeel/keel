import { models, CreatePostWithRole } from "@teamkeel/sdk";

export default CreatePostWithRole(async (_, inputs) => {
  return models.post.create({
    title: inputs.title,
    businessId: inputs.business.id,
  });
});
