import { DeletePost } from "@teamkeel/sdk";

export default DeletePost(async (inputs, api, ctx) => {
  const post = await api.models.post.delete({ id: inputs.id });
  return post;
});
