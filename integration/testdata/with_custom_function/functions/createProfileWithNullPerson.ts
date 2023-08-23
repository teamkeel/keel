import { CreateProfileWithNullPerson } from "@teamkeel/sdk";

export default CreateProfileWithNullPerson({
  // we use ts-ignore below to disable typechecking for this test case
  // because we want to ensure that our runtime does not accept null as a value
  // for this non-nullable field
  // @ts-ignore
  beforeWrite: async (ctx, inputs) => {
    return {
      personId: null,
    };
  },
});
