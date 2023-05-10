import { ListBooksByPublisherName } from "@teamkeel/sdk";

export default ListBooksByPublisherName(async (ctx, inputs, api) => {
  api.permissions.allow();

  return api.models.book.findMany({
    author: {
      publisher: {
        name: inputs.where?.authorPublisherName,
      },
    },
  });
});
