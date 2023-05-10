import { ListPosts } from "@teamkeel/sdk";

export default ListPosts(async (_, inputs, api) => {
  const result = await api.models.post.findMany({});

  return result;
});
