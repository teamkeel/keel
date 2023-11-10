import { UpdateBookBeforeQuery } from "@teamkeel/sdk";

export default UpdateBookBeforeQuery({
  beforeQuery(ctx, inputs, query) {
    query = query.where({
      published: true,
    });

    if (inputs.where.returnRecord) {
      return query.findOne();
    }

    return query;
  },
});
