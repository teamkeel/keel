import { DeleteBookNoInputs, DeleteBookNoInputsHooks } from "@teamkeel/sdk";

// To learn more about what you can do with hooks, visit https://docs.keel.so/functions
const hooks: DeleteBookNoInputsHooks = {};

export default DeleteBookNoInputs({
  beforeQuery(ctx, query) {
    return query.where({
      title: "The Farseer 2",
    });
  },
});
