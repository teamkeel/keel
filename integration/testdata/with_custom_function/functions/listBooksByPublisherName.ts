import { models, ListBooksByPublisherName } from "@teamkeel/sdk";

export default ListBooksByPublisherName(async (ctx, inputs) => {
  return models.book.findMany({
    author: {
      publisher: {
        name: inputs.where?.authorPublisherName,
      },
    },
  });
});
