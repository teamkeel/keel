import { DeleteBookAfterWrite, models } from "@teamkeel/sdk";

export default DeleteBookAfterWrite({
  async afterWrite(ctx, inputs, data) {
    await models.deletedBook.create({
      bookId: data.id,
      title: `${data.title} (${inputs.reason})`,
      deletedAt: ctx.now(),
    });
  },
});
