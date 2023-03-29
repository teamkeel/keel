import { GetPost } from "@teamkeel/sdk";

export default GetPost(async (inputs, api, ctx) => {
  const result = await api.models.post.findOne({ id: inputs.id });

  return result;
});
