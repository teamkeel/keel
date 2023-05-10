import { GetPublisherByBook } from "@teamkeel/sdk";

export default GetPublisherByBook(async (ctx, inputs, api) => {
  api.permissions.allow();

  return api.models.publisher.findOne({
    authors: {
      books: {
        id: inputs.bookId,
      },
    },
  });
});
