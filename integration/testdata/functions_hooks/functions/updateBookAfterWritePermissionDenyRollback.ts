import { UpdateBookAfterWritePermissionDenyRollback, models, permissions } from "@teamkeel/sdk";

export default UpdateBookAfterWritePermissionDenyRollback({
  async beforeWrite(ctx, inputs, values, record) {
    // Create a bookUpdates record that should be rolled back when permission is denied in afterWrite
    await models.bookUpdates.create({
      bookId: record.id,
      updateCount: 777,
    });

    return {
      ...values,
      title: values.title.toUpperCase(),
    };
  },
  async afterWrite(ctx, inputs, data) {
    // Create a review that should also be rolled back
    await models.review.create({
      bookId: data.id,
      review: "afterWrite review - " + data.title,
    });

    permissions.deny();
  },
});
