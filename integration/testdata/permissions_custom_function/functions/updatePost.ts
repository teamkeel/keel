import { UpdatePost } from "@teamkeel/sdk";
import { PrismaClient } from "@prisma/client";
const models = new PrismaClient();

export default UpdatePost(async (_, inputs) => {
  const post = await models.post.update({
    data: inputs.values,
    where: inputs.where,
  });
  return post;
});
