import { CreatePersonWithEnvVar } from "@teamkeel/sdk";

export default CreatePersonWithEnvVar({
  beforeWrite: async (ctx, inputs) => {
    return {
      ...inputs,
      name: ctx.env.TEST,
    };
  },
});
