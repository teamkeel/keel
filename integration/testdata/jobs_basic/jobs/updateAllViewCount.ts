import { UpdateAllViewCount, models } from "@teamkeel/sdk";

export default UpdateAllViewCount(async (ctx) => {
  const posts = await models.post.findMany({});

  for (const post of posts) {
    const views = await models.postViews.findMany({
      where: { postId: post!.id },
    });

    let totalViewCount = 0;
    views.forEach(function (v) {
      totalViewCount += v.views;
    });

    await models.post.update(
      { id: post!.id },
      {
        viewCount: totalViewCount,
        viewCountUpdated: ctx.now(),
      }
    );
  }
});
