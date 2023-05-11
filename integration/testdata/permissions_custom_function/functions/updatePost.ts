import { models, UpdatePost } from "@teamkeel/sdk";

export default UpdatePost(async (_, inputs) => {
  const post = await models.post.update(inputs.where, inputs.values);
  return post;
});
