import { UpdateNotPermittedFn, models } from "@teamkeel/sdk";

export default UpdateNotPermittedFn(async (ctx, inputs) => {
  const book = await models.book.update(inputs.where, inputs.values);
  return book;
});
