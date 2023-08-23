import { GetPublisherByBook } from "@teamkeel/sdk";

export default GetPublisherByBook({
  beforeQuery: (ctx, inputs, query) => {
    return query
      .where({
        authors: {
          books: {
            id: inputs.bookId,
          },
        },
      })
      .findOne();
  },
});
