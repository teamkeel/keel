import { GetPost } from "@teamkeel/sdk";
import { PrismaClient } from "@prisma/client";
const models = new PrismaClient();

export default GetPost(async (_, inputs) => {
  const result = await models.post.findFirst({ where: { id: inputs.id } });
  return result;
});
