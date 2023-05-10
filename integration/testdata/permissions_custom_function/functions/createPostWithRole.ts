import { CreatePostWithRole } from "@teamkeel/sdk";

export default CreatePostWithRole(async (_, inputs, api) => {
  return api.models.post.create({
    title: inputs.title,
    businessId: inputs.business.id,
  });
});
