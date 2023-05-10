import { GetSecretPost } from "@teamkeel/sdk";

// shh
export default GetSecretPost(async (_, inputs, api) => {
  const result = await api.models.post.findOne({ id: inputs.id });

  return result;
});
