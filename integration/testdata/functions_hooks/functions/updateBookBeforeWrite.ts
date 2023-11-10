import { UpdateBookBeforeWrite, permissions } from "@teamkeel/sdk";

export default UpdateBookBeforeWrite({
  async beforeWrite(ctx, inputs, values, record) {
    if (values.title.toLowerCase().includes("how to build a bomb")) {
      permissions.deny();
    }

    return {
      ...values,
      published: !record.published,
    };
  },
});
