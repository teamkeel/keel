import { DeletePost } from "@teamkeel/sdk";
import { PrismaClient } from "@prisma/client";
const models = new PrismaClient();

export default DeletePost(async (_, inputs) => {
  const post = await models.post.delete({ where: inputs });
  return post.id;
});
