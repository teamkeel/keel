import { DeleteBookBeforeQueryReturnRecord } from "@teamkeel/sdk";

export default DeleteBookBeforeQueryReturnRecord({
  async beforeQuery(ctx, inputs, query) {
    const record = await query.findOne();
    return record;
  },
});
