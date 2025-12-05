import { CreateBookBeforeWriteExceptionRollback, models } from "@teamkeel/sdk";

export default CreateBookBeforeWriteExceptionRollback({
  async beforeWrite(ctx, inputs, values) {
    // Create a deletedBook record that should be rolled back when the exception is thrown
    // Using deletedBook because it stores bookId as a plain ID field without FK constraint
    await models.deletedBook.create({
      bookId: "00000000-0000-0000-0000-000000000000",
      title: "orphaned record - " + values.title,
      deletedAt: ctx.now(),
    });

    throw new Error("exception in beforeWrite");
  },
});
