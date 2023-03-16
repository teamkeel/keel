import { ListPosts } from "@teamkeel/sdk";

export default ListPosts(async (inputs, api, ctx) => {
  const result = await api.models.post.findMany({});

  return result;
});
