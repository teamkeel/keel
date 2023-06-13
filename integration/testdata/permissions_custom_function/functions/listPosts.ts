import { ListPosts } from "@teamkeel/sdk";
import { PrismaClient } from "@prisma/client";
const models = new PrismaClient();

export default ListPosts(async (_, inputs) => {
  const result = await models.post.findMany();

  return result;
});
