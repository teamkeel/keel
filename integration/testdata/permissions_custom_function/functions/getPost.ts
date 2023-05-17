import { models, GetPost } from "@teamkeel/sdk";

export default GetPost(async (_, inputs) => {
  const result = await models.post.findOne({ id: inputs.id });
  return result;
});
