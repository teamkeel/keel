import { ListBooksBeforeQuery } from "@teamkeel/sdk";

// This function is testing that the beforeQuery hook of a
// list function can return a mutated version of the provided QueryBuilder
export default ListBooksBeforeQuery({
  beforeQuery(ctx, inputs, query) {
    return query.where({
      title: {
        endsWith: "Magic",
      },
    });
  },
});
