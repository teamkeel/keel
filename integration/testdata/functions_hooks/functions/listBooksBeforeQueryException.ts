import { ListBooksBeforeQueryException } from "@teamkeel/sdk";

export default ListBooksBeforeQueryException({
  beforeQuery(ctx, inputs, query) {
    throw new Error("exception in list beforeQuery");
  },
});
