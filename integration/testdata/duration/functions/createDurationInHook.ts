import { CreateDurationInHook, CreateDurationInHookHooks, Duration } from '@teamkeel/sdk';

// To learn more about what you can do with hooks, visit https://docs.keel.so/functions
const hooks: CreateDurationInHookHooks = {
    beforeWrite: async (ctx, inputs) => {
        return {
            dur: Duration.fromISOString("PT1H")
        }
    },
};

export default CreateDurationInHook(hooks);
	