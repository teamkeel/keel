import { GetPermittedFn, models } from "@teamkeel/sdk";

export default GetPermittedFn(async (ctx, inputs) => {
  const book = await models.book.findOne(inputs);
  return book;
});
