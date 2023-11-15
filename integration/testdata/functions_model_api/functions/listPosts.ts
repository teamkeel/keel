import { models, ListPosts } from "@teamkeel/sdk";

export default ListPosts({
  beforeQuery: async (ctx, inputs, query) => {
    const { where } = inputs;

    return models.post.findMany({
      orderBy: where?.orderBy
        ? {
            [where.orderBy]: where.sortOrder,
          }
        : undefined,
      limit: where?.limit ? where.limit : undefined,
      offset: where?.offset ? where.offset : undefined,
    });
  },
});
