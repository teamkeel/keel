import { CreatePost } from "@teamkeel/sdk";
import { PrismaClient } from "@prisma/client";
const models = new PrismaClient();
export default CreatePost(async (_, inputs) => {
  const result = await models.post.create({
    data: {
      id: "123",
      createdAt: new Date(),
      updatedAt: new Date(),
      title: inputs.title,
      businessId: inputs.business.id,
    },
  });

  return result;
});
