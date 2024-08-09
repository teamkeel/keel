import {
  errors,
  HookNotFoundCustomMessage,
  HookNotFoundCustomMessageHooks,
} from "@teamkeel/sdk";

// To learn more about what you can do with hooks, visit https://docs.keel.so/functions
const hooks: HookNotFoundCustomMessageHooks = {
  beforeQuery(ctx, inputs, query) {
    throw new errors.NotFound("nothing here");
    return query;
  },
};

export default HookNotFoundCustomMessage(hooks);
