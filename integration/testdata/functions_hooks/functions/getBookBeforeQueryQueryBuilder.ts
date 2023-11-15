import { GetBookBeforeQueryQueryBuilder } from "@teamkeel/sdk";

// This function is testing that a beforeQuery hook in a get function can return a QueryBuilder
export default GetBookBeforeQueryQueryBuilder({
  beforeQuery(ctx, inputs, query) {
    if (!inputs.allowUnpublished) {
      return query.where({
        published: true,
      });
    }
    return query;
  },
});
