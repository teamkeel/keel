import { GetPostWithHook, models } from "@teamkeel/sdk";

export default GetPostWithHook({
  async beforeQuery(ctx, inputs, query) {
    // Log that this hook was executed
    await models.hookLog.create({
      actionName: "getPostWithHook",
      hookName: "beforeQuery",
      executedAt: ctx.now(),
    });

    return query;
  },
  async afterQuery(ctx, inputs, data) {
    // Log that this hook was executed
    await models.hookLog.create({
      actionName: "getPostWithHook",
      hookName: "afterQuery",
      executedAt: ctx.now(),
    });

    return data;
  },
});
