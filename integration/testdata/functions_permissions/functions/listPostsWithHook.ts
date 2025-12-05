import { ListPostsWithHook, models } from "@teamkeel/sdk";

export default ListPostsWithHook({
  async beforeQuery(ctx, inputs, query) {
    // Log that this hook was executed
    await models.hookLog.create({
      actionName: "listPostsWithHook",
      hookName: "beforeQuery",
      executedAt: ctx.now(),
    });

    return query;
  },
  async afterQuery(ctx, inputs, data) {
    // Log that this hook was executed
    await models.hookLog.create({
      actionName: "listPostsWithHook",
      hookName: "afterQuery",
      executedAt: ctx.now(),
    });

    return data;
  },
});
