import { models, DeletePost } from "@teamkeel/sdk";

export default DeletePost(async (_, inputs) => {
  const post = await models.post.delete(inputs);
  return post;
});
