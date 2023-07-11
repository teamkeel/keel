import { UpdateGoldStarFromEnv, models, Status } from "@teamkeel/sdk";

export default UpdateGoldStarFromEnv(async (ctx) => {
  const posts = await models.post.findMany({});

  for (const post of posts) {
    if (post!.viewCount > parseInt(ctx.env.GOLD_STAR)) {
      await models.post.update(
        { id: post!.id },
        {
          status: Status.GoldPost,
        }
      );
    }
  }
});
