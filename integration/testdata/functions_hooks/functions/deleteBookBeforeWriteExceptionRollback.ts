import { DeleteBookBeforeWriteExceptionRollback, models } from "@teamkeel/sdk";

export default DeleteBookBeforeWriteExceptionRollback({
  async beforeWrite(ctx, inputs, data) {
    // Create a deletedBook record that should be rolled back when the exception is thrown
    await models.deletedBook.create({
      bookId: data.id,
      title: "to be rolled back - " + data.title,
      deletedAt: ctx.now(),
    });

    throw new Error("exception in delete beforeWrite");
  },
});
