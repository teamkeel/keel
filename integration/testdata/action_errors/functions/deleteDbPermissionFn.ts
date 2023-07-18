import { DeleteDbPermissionFn, models } from "@teamkeel/sdk";

export default DeleteDbPermissionFn(async (ctx, inputs) => {
  const book = await models.book.delete(inputs);
  return book;
});
