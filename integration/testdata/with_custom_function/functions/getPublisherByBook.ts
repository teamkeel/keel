import { models, GetPublisherByBook } from "@teamkeel/sdk";

export default GetPublisherByBook(async (ctx, inputs) => {
  return models.publisher.findOne({
    authors: {
      books: {
        id: inputs.bookId,
      },
    },
  });
});
