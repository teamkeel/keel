import { DeleteBookAfterWriteExceptionRollback, models } from "@teamkeel/sdk";

export default DeleteBookAfterWriteExceptionRollback({
  async beforeWrite(ctx, inputs, data) {
    // Create a deletedBook record that should be rolled back when exception is thrown in afterWrite
    await models.deletedBook.create({
      bookId: data.id,
      title: "beforeWrite - " + data.title,
      deletedAt: ctx.now(),
    });
  },
  async afterWrite(ctx, inputs, data) {
    // Create another deletedBook record that should also be rolled back
    await models.deletedBook.create({
      bookId: data.id,
      title: "afterWrite - " + data.title,
      deletedAt: ctx.now(),
    });

    throw new Error("exception in delete afterWrite");
  },
});
