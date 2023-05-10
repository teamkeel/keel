import { UpdatePost } from "@teamkeel/sdk";

export default UpdatePost(async (_, inputs, api) => {
  const post = await api.models.post.update(inputs.where, inputs.values);
  return post;
});
