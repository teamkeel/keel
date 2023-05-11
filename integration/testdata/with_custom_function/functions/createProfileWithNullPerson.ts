import { models, CreateProfileWithNullPerson } from "@teamkeel/sdk";

export default CreateProfileWithNullPerson((ctx, inputs) => {
  return models.profile.create({
    // Given the create method is type we actually have to bypass
    // TypeScript to pass null here and get the error we want
    // @ts-ignore
    personId: null,
  });
});
