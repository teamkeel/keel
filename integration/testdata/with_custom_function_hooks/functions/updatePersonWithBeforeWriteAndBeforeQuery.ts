import {
  UpdatePersonWithBeforeWriteAndBeforeQuery,
  UpdatePersonWithBeforeWriteAndBeforeQueryHooks,
  models,
} from "@teamkeel/sdk";

// To learn more about what you can do with hooks,
// visit https://docs.keel.so/functions
const hooks: UpdatePersonWithBeforeWriteAndBeforeQueryHooks = {
  beforeWrite: async (ctx, inputs, values) => {
    return {
      ...values,
      title: values.title.repeat(2),
    };
  },
  beforeQuery: async (ctx, { where }, values) => {
    const updated = await models.person.update(where, values);

    return {
      ...updated,
      title: values.title.repeat(2),
    };
  },
};

export default UpdatePersonWithBeforeWriteAndBeforeQuery(hooks);
