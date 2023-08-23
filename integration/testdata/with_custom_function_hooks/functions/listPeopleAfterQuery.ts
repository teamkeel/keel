import {
  ListPeopleAfterQuery,
  ListPeopleAfterQueryHooks,
  models,
} from "@teamkeel/sdk";

// To learn more about what you can do with hooks,
// visit https://docs.keel.so/functions
const hooks: ListPeopleAfterQueryHooks = {
  afterQuery: async (ctx, inputs, records) => {
    await models.log.create({
      msg: `List results: ${records.length}`,
    });

    return records;
  },
};

export default ListPeopleAfterQuery(hooks);
