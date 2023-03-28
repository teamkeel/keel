import { GetSecretPost } from "@teamkeel/sdk";

// shh
export default GetSecretPost(async (inputs, api, ctx) => {
  const result = await api.models.post.findOne({ id: inputs.id });

  return result;
});
