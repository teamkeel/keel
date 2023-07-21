import { CreateDbPermissionFn, models } from "@teamkeel/sdk";

export default CreateDbPermissionFn(async (ctx, inputs) => {
  return await models.book.create({
    title: inputs.title,
    lastUpdatedById: inputs.lastUpdatedBy!.id!,
  });
});
