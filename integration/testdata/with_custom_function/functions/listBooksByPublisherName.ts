import { models, ListBooksByPublisherName } from "@teamkeel/sdk";

export default ListBooksByPublisherName(async (ctx, inputs) => {
  return models.book.findMany({
    where: {
      author: {
        publisher: {
          name: inputs.where.author.publisher?.name,
        },
      },
    },
  });
});
