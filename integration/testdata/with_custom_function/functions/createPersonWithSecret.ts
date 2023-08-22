import { CreatePersonWithSecret } from "@teamkeel/sdk";

export default CreatePersonWithSecret({
  beforeWrite: async (ctx, inputs) => {
    return {
      ...inputs,
      name: ctx.secrets.NAME_API_KEY,
    };
  },
});
