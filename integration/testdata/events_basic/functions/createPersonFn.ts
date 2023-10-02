import { CreatePersonFn, CreatePersonFnHooks } from "@teamkeel/sdk";

const hooks: CreatePersonFnHooks = {
  afterWrite: async (ctx, inputs, person) => {
    if (inputs.name === "") {
      throw new Error("error occurred");
    }
  },
};

export default CreatePersonFn(hooks);
