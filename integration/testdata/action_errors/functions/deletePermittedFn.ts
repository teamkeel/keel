import { DeletePermittedFn, models } from "@teamkeel/sdk";

export default DeletePermittedFn(async (ctx, inputs) => {
  const book = await models.book.delete(inputs);
  return book;
});
