import { UpdatePermittedFn, models } from "@teamkeel/sdk";

export default UpdatePermittedFn(async (ctx, inputs) => {
  const book = await models.book.update(inputs.where, {
    title: inputs.values.title,
    lastUpdatedById: inputs.values.lastUpdatedBy?.id,
  });
  return book;
});
