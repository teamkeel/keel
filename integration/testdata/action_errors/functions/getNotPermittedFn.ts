import { GetNotPermittedFn, models } from "@teamkeel/sdk";

export default GetNotPermittedFn(async (ctx, inputs) => {
  const book = await models.book.findOne(inputs);
  return book;
});
