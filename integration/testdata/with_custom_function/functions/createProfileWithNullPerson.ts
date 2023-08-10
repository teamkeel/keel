import { CreateProfileWithNullPerson } from "@teamkeel/sdk";

export default CreateProfileWithNullPerson({
  // @ts-ignore
  beforeWrite: async (ctx, inputs, values) => {
    return {
      personId: null,
    };
  },
});
