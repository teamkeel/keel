import { GetPost } from "@teamkeel/sdk";

export default GetPost(async (_, inputs, api) => {
  const result = await api.models.post.findOne({ id: inputs.id });

  return result;
});
