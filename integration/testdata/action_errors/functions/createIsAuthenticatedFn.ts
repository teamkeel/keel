import { CreateIsAuthenticatedFn, models } from "@teamkeel/sdk";

export default CreateIsAuthenticatedFn(async (ctx, inputs) => {
  return await models.book.create({
    title: inputs.title,
    lastUpdatedById: ctx.identity!.id,
  });
});
