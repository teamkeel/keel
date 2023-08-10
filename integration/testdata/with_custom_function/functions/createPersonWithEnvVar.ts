import { CreatePersonWithEnvVar } from "@teamkeel/sdk";

export default CreatePersonWithEnvVar({
  beforeWrite: async (ctx, inputs, values) => {
    return {
      ...inputs,
      name: ctx.env.TEST,
    };
  },
});
