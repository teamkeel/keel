import { DeleteNotPermittedFn, models } from "@teamkeel/sdk";

export default DeleteNotPermittedFn(async (ctx, inputs) => {
  const book = await models.book.delete(inputs);
  return book;
});
