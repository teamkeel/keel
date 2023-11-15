import { UpdateBookAfterWrite, models } from "@teamkeel/sdk";

export default UpdateBookAfterWrite({
  async afterWrite(ctx, inputs, data) {
    const updates = await models.bookUpdates.findOne({
      bookId: data.id,
    });

    if (!updates) {
      await models.bookUpdates.create({
        bookId: data.id,
        updateCount: 1,
      });
    } else {
      await models.bookUpdates.update(
        {
          id: updates.id,
        },
        {
          updateCount: updates.updateCount + 1,
        }
      );
    }

    return {
      ...data,
      title: data.title.toUpperCase(),
    };
  },
});
