import { ListPostsByDateWithHook, ListPostsByDateWithHookHooks } from '@teamkeel/sdk';

// To learn more about what you can do with hooks, visit https://docs.keel.so/functions
const hooks : ListPostsByDateWithHookHooks = {
    beforeQuery(ctx, inputs, query) {
        return query.where({
            aDate: {
                beforeRelative: "today",
            },
        });
    },
};


export default ListPostsByDateWithHook(hooks);
	