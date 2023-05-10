import { DeletePost } from "@teamkeel/sdk";

export default DeletePost(async (_, inputs, api) => {
  const post = await api.models.post.delete(inputs);
  return post;
});
