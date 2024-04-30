import { ReadPostsByTeam, models } from "@teamkeel/sdk";

export default ReadPostsByTeam(async (_, inputs) => {
  const posts = await models.post
    .where({
      identity: {
        teamId: { equals: inputs.team },
      },
    })
    .findMany();

  return posts;
});
