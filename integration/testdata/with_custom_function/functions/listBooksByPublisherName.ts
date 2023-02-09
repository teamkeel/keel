import { ListBooksByPublisherName } from "@teamkeel/sdk";

export default ListBooksByPublisherName(async (inputs, api) => {
  return api.models.book.findMany({
    author: {
      publisher: {
        name: {
          equals: inputs.where.authorPublisherName,
        },
      },
    },
  });
});
