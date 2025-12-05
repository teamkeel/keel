import { CreatePostWithHook, models } from "@teamkeel/sdk";

export default CreatePostWithHook({
  async beforeWrite(ctx, inputs, values) {
    // Log that this hook was executed
    await models.hookLog.create({
      actionName: "createPostWithHook",
      hookName: "beforeWrite",
      executedAt: ctx.now(),
    });

    return values;
  },
  async afterWrite(ctx, inputs, data) {
    // Log that this hook was executed
    await models.hookLog.create({
      actionName: "createPostWithHook",
      hookName: "afterWrite",
      executedAt: ctx.now(),
    });

    return data;
  },
});
