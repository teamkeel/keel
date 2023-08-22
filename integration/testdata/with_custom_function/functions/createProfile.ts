import { CreateProfile } from "@teamkeel/sdk";

export default CreateProfile({
  beforeWrite: async (ctx, inputs) => {
    return { personId: inputs.person.id };
  },
});
