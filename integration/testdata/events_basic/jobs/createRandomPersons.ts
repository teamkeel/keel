import { models, CreateRandomPersons } from "@teamkeel/sdk";

export default CreateRandomPersons(async (ctx, inputs) => {
  await models.person.create({ name: "Keelson", email: "keelson@keel.xyz" });
  await models.person.create({ name: "Weaveton", email: "weaveton@keel.xyz" });

  if (inputs.raiseException) {
    throw new Error("error occurred");
  }
});
