import {
  CreatePersonUsingHook,
  CreatePersonUsingHookHooks,
} from "@teamkeel/sdk";

const hooks: CreatePersonUsingHookHooks = {};

export default CreatePersonUsingHook({
  beforeWrite: async (ctx, _) => {
    return {
      name: ctx.headers.get("Person-Name")!,
    };
  },
});
