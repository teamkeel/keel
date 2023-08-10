import { CreateProfile } from "@teamkeel/sdk";

export default CreateProfile({
  beforeWrite: async (ctx, inputs, values) => {
    return { personId: inputs.person.id };
  },
});
