import { CreateBookNoInputs, CreateBookNoInputsHooks } from "@teamkeel/sdk";
import { createExpect } from "vitest";

// To learn more about what you can do with hooks, visit https://docs.keel.so/functions
const hooks: CreateBookNoInputsHooks = {};

export default CreateBookNoInputs({
  beforeWrite(ctx, values) {
    return {
      ...values,
      title: "The Farseer",
    };
  },
  afterWrite(ctx, data) {
    if (data.title !== "The Farseer") {
      return new Error("Title must be 'The Farseer'");
    }
    return {
      ...data,
      title: "The Farseer 2",
    };
  },
});
