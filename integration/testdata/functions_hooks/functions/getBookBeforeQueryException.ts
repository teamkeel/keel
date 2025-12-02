import { GetBookBeforeQueryException } from "@teamkeel/sdk";

export default GetBookBeforeQueryException({
  beforeQuery(ctx, inputs, query) {
    throw new Error("exception in get beforeQuery");
  },
});
