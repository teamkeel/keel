import { CreatePermittedFn, models } from "@teamkeel/sdk";

export default CreatePermittedFn(async (ctx, inputs) => {
  return await models.book.create({
    title: inputs.title,
    lastUpdatedById: inputs.lastUpdatedBy!.id!,
  });
});
