import { CreateBookAfterWrite, models } from "@teamkeel/sdk";

export default CreateBookAfterWrite({
  async afterWrite(ctx, inputs, data) {
    await models.review.create({
      bookId: data.id,
      review: inputs.review,
    });
  },
});
