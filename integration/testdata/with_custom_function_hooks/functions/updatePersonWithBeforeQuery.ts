import {
  UpdatePersonWithBeforeQuery,
  UpdatePersonWithBeforeQueryHooks,
  models,
} from "@teamkeel/sdk";

// To learn more about what you can do with hooks,
// visit https://docs.keel.so/functions
const hooks: UpdatePersonWithBeforeQueryHooks = {
  beforeQuery: async (ctx, inputs) => {
    const updatedRecord = await models.person.update(
      {
        id: "xxx",
      },
      inputs.values
    );

    return updatedRecord;
  },
};

export default UpdatePersonWithBeforeQuery(hooks);
