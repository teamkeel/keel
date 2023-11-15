import { DeleteBookBeforeQuery } from "@teamkeel/sdk";

export default DeleteBookBeforeQuery({
  async beforeQuery(ctx, inputs, query) {
    if (!inputs.allowPublished) {
      return query.where({
        published: false,
      });
    }

    return query;
  },
});
