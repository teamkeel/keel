import { CreatePostWithRole } from "@teamkeel/sdk";
import { PrismaClient } from "@prisma/client";
const models = new PrismaClient();

export default CreatePostWithRole(async (_, inputs) => {
  return models.post.create({
    data: {
      title: inputs.title,
      businessId: inputs.business.id,
      id: "123",
      createdAt: new Date(),
      updatedAt: new Date(),
    },
  });
});
