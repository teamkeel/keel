import { GetSecretPost } from "@teamkeel/sdk";
import { PrismaClient } from "@prisma/client";
const models = new PrismaClient();

// shh
export default GetSecretPost(async (_, inputs) => {
  const result = await models.post.findFirst({ where: { id: inputs.id } });

  return result;
});
