import { models, GetSecretPost } from "@teamkeel/sdk";

// shh
export default GetSecretPost(async (_, inputs) => {
  const result = await models.post.findOne({ id: inputs.id });

  return result;
});
