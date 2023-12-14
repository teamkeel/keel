import { CreateAuthorAndBooks } from "@teamkeel/sdk";

export default CreateAuthorAndBooks({
  beforeWrite(ctx, inputs, values) {
    return {
      ...values,
      books: inputs.books.map((b) => {
        return {
          ...b,
          published: true,
        };
      }),
    };
  },
});
