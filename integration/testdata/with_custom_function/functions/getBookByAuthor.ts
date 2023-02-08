import { GetBookByAuthor } from "@teamkeel/sdk";

export default GetBookByAuthor(async (inputs, api) => {
  return api.models.book.findOne({
    author: {
      id: inputs.authorId,
    },
  });
});
