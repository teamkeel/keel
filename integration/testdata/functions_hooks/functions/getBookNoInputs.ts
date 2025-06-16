import { GetBookNoInputs, GetBookNoInputsHooks } from "@teamkeel/sdk";

// To learn more about what you can do with hooks, visit https://docs.keel.so/functions
const hooks: GetBookNoInputsHooks = {};

export default GetBookNoInputs({
  beforeQuery(ctx, inputs, query) {
    return query.where({
      title: "The Farseer",
    });
  },
  afterQuery(ctx, inputs, data) {
    // if (data.title !== "The Farseer") {
    //   return new Error("Title must be 'The Farseer'");
    // }
    return {
      ...data,
      title: "The Farseer 2",
    };
  },
});
