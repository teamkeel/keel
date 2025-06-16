import { UpdateBookNoInputs, UpdateBookNoInputsHooks } from "@teamkeel/sdk";

// To learn more about what you can do with hooks, visit https://docs.keel.so/functions
const hooks: UpdateBookNoInputsHooks = {};

export default UpdateBookNoInputs({
  beforeWrite(ctx, inputs, values) {
    return {
      ...values,
      title: "The Farseer 2",
    };
  },
  afterWrite(ctx, inputs, data) {
    return {
      ...data,
      title: "The Farseer",
    };
  },
});
