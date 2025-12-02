import {
  CreateBookAfterWritePermissionDenyRollback,
  models,
  permissions,
} from "@teamkeel/sdk";

export default CreateBookAfterWritePermissionDenyRollback({
  async beforeWrite(ctx, inputs, values) {
    // Create a deletedBook record that should be rolled back when permission is denied in afterWrite
    // Using deletedBook because it stores bookId as a plain ID field without FK constraint
    await models.deletedBook.create({
      bookId: "00000000-0000-0000-0000-000000000000",
      title: "beforeWrite record - " + values.title,
      deletedAt: ctx.now(),
    });

    return values;
  },
  async afterWrite(ctx, inputs, data) {
    // Create a review that should also be rolled back (now the book exists so we can use bookId)
    await models.review.create({
      bookId: data.id,
      review: "afterWrite review - " + data.title,
    });

    permissions.deny();
  },
});
