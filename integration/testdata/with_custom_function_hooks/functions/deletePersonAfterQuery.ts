import {
  DeletePersonAfterQuery,
  DeletePersonAfterQueryHooks,
  models,
} from "@teamkeel/sdk";

// To learn more about what you can do with hooks,
// visit https://docs.keel.so/functions
const hooks: DeletePersonAfterQueryHooks = {
  afterQuery: async (ctx, inputs, id) => {
    await models.log.create({
      msg: `deleted person ${id}`,
    });

    return id;
  },
};

export default DeletePersonAfterQuery(hooks);
