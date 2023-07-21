import { CreateNotPermittedFn, models } from "@teamkeel/sdk";

export default CreateNotPermittedFn(async (ctx, inputs) => {
  return await models.book.create({
    title: inputs.title,
    lastUpdatedById: inputs.lastUpdatedBy?.id!,
  });
});
