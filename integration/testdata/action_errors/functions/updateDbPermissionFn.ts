import { UpdateDbPermissionFn, models } from "@teamkeel/sdk";

export default UpdateDbPermissionFn(async (ctx, inputs) => {
  const book = await models.book.update(inputs.where, inputs.values);
  return book;
});
