import { UpdateBookBeforeWriteExceptionRollback, models } from "@teamkeel/sdk";

export default UpdateBookBeforeWriteExceptionRollback({
  async beforeWrite(ctx, inputs, values, record) {
    // Create a bookUpdates record that should be rolled back when the exception is thrown
    await models.bookUpdates.create({
      bookId: record.id,
      updateCount: 999,
    });

    throw new Error("exception in update beforeWrite");
  },
});
