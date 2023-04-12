import { CreatePostWithRole } from "@teamkeel/sdk";

export default CreatePostWithRole(async (inputs, api, ctx) => {
  return api.models.post.create({
    title: inputs.title,
    businessId: inputs.business.id,
  });
});
