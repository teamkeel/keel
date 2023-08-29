import {
  UpdatePersonWithBeforeWriteAndBeforeQuery,
  UpdatePersonWithBeforeWriteAndBeforeQueryHooks,
  models,
} from "@teamkeel/sdk";

// To learn more about what you can do with hooks,
// visit https://docs.keel.so/functions
const hooks: UpdatePersonWithBeforeWriteAndBeforeQueryHooks = {
  beforeWrite: async (ctx, inputs) => {
    return {
      ...inputs.values,
      title: inputs.values.title.repeat(2),
    };
  },
  beforeQuery: async (ctx, inputs) => {
    const updated = await models.person.update(inputs.where, inputs.values);

    return {
      ...updated,
      title: updated.title.repeat(2),
    };
  },
};

export default UpdatePersonWithBeforeWriteAndBeforeQuery(hooks);
