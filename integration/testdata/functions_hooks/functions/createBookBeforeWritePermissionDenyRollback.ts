import {
  CreateBookBeforeWritePermissionDenyRollback,
  models,
  permissions,
} from "@teamkeel/sdk";

export default CreateBookBeforeWritePermissionDenyRollback({
  async beforeWrite(ctx, inputs, values) {
    // Create a deletedBook record that should be rolled back when permission is denied
    // Using deletedBook because it stores bookId as a plain ID field without FK constraint
    await models.deletedBook.create({
      bookId: "00000000-0000-0000-0000-000000000000",
      title: "orphaned record - " + values.title,
      deletedAt: ctx.now(),
    });

    permissions.deny();
  },
});
