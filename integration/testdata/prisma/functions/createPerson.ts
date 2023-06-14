import { PrismaClient } from "@prisma/client";
import { CreatePerson } from "@teamkeel/sdk";

const models = new PrismaClient();

export default CreatePerson(async (ctx, inputs) => {
  return models.person.create({
    data: {
      ...inputs,
      id: "extremely-unique",
      createdAt: new Date(),
      updatedAt: new Date(),
    },
  });
});
