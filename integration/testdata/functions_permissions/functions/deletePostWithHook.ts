import { DeletePostWithHook, models } from "@teamkeel/sdk";

export default DeletePostWithHook({
  async beforeQuery(ctx, inputs, query) {
    // Log that this hook was executed
    await models.hookLog.create({
      actionName: "deletePostWithHook",
      hookName: "beforeQuery",
      executedAt: ctx.now(),
    });

    return query;
  },
  async beforeWrite(ctx, inputs, data) {
    // Log that this hook was executed
    await models.hookLog.create({
      actionName: "deletePostWithHook",
      hookName: "beforeWrite",
      executedAt: ctx.now(),
    });
  },
  async afterWrite(ctx, inputs, data) {
    // Log that this hook was executed
    await models.hookLog.create({
      actionName: "deletePostWithHook",
      hookName: "afterWrite",
      executedAt: ctx.now(),
    });
  },
});
