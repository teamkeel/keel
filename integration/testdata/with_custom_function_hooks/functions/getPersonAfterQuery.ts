import {
  GetPersonAfterQuery,
  GetPersonAfterQueryHooks,
  models,
} from "@teamkeel/sdk";

// To learn more about what you can do with hooks,
// visit https://docs.keel.so/functions
const hooks: GetPersonAfterQueryHooks = {
  afterQuery: async (ctx, inputs, person) => {
    await models.log.create({
      msg: `Fetched ${person.id}`,
    });

    return person;
  },
};

export default GetPersonAfterQuery(hooks);
