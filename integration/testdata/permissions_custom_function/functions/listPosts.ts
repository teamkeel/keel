import { models, ListPosts } from "@teamkeel/sdk";

export default ListPosts(async (_, inputs) => {
  const result = await models.post.findMany({});

  return result;
});
