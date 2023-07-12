import { UpdateViewCount, models } from "@teamkeel/sdk";

const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms));

export default UpdateViewCount(async (ctx, inputs) => {
  const post = await models.post.findOne({ id: inputs.postId });
  if (post == null) {
    return;
  }

  const views = await models.postViews.findMany({
    where: { postId: post!.id },
  });

  let totalViewCount = 0;
  views.forEach(function (v) {
    totalViewCount += v.views;
  });

  await sleep(1000); // Tests that jobs are being awaited correctly in tests

  await models.post.update(
    { id: post!.id },
    {
      viewCount: totalViewCount,
      viewCountUpdated: ctx.now(),
    }
  );
});
