import { CreateProfileWithNullPerson } from "@teamkeel/sdk";

export default CreateProfileWithNullPerson((inputs, api) => {
  api.permissions.allow();

  return api.models.profile.create({
    // Given the create method is type we actually have to bypass
    // TypeScript to pass null here and get the error we want
    // @ts-ignore
    personId: null,
  });
});
