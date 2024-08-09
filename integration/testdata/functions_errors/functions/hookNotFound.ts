import { errors, HookNotFound, HookNotFoundHooks } from "@teamkeel/sdk";

// To learn more about what you can do with hooks, visit https://docs.keel.so/functions
const hooks: HookNotFoundHooks = {
  beforeQuery(ctx, inputs, query) {
    throw new errors.NotFound();
    return query;
  },
};

export default HookNotFound(hooks);
