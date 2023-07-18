import { GetDbPermissionFn, models } from "@teamkeel/sdk";

export default GetDbPermissionFn(async (ctx, inputs) => {
  const book = await models.book.findOne(inputs);
  return book;
});
