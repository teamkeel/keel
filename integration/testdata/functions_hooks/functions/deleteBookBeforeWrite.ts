import { DeleteBookBeforeWrite, models, permissions } from "@teamkeel/sdk";

export default DeleteBookBeforeWrite({
  async beforeWrite(ctx, inputs, data) {
    if (!inputs.allowPublished && data.published) {
      permissions.deny();
    }

    await models.deletedBook.create({
      bookId: data.id,
      title: data.title,
      deletedAt: ctx.now(),
    });
  },
});
