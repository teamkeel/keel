import { UpdatePostWithHook, models } from "@teamkeel/sdk";

export default UpdatePostWithHook({
  async beforeQuery(ctx, inputs, query) {
    // Log that this hook was executed
    await models.hookLog.create({
      actionName: "updatePostWithHook",
      hookName: "beforeQuery",
      executedAt: ctx.now(),
    });

    return query;
  },
  async beforeWrite(ctx, inputs, values, record) {
    // Log that this hook was executed
    await models.hookLog.create({
      actionName: "updatePostWithHook",
      hookName: "beforeWrite",
      executedAt: ctx.now(),
    });

    return values;
  },
  async afterWrite(ctx, inputs, data) {
    // Log that this hook was executed
    await models.hookLog.create({
      actionName: "updatePostWithHook",
      hookName: "afterWrite",
      executedAt: ctx.now(),
    });

    return data;
  },
});
