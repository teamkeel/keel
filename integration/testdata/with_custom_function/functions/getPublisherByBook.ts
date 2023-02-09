import { GetPublisherByBook } from "@teamkeel/sdk";

export default GetPublisherByBook(async (inputs, api) => {
  return api.models.publisher.findOne({
    authors: {
      books: {
        id: inputs.bookId,
      },
    },
  });
});
