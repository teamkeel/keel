import { UpdateBookAfterWriteExceptionRollback, models } from "@teamkeel/sdk";

export default UpdateBookAfterWriteExceptionRollback({
  async beforeWrite(ctx, inputs, values, record) {
    // Create a bookUpdates record that should be rolled back when exception is thrown in afterWrite
    await models.bookUpdates.create({
      bookId: record.id,
      updateCount: 888,
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

    throw new Error("exception in update afterWrite");
  },
});
