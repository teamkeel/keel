import {
  GetPersonBeforeQuery,
  GetPersonBeforeQueryHooks,
  Sex,
} from "@teamkeel/sdk";

// To learn more about what you can do with hooks,
// visit https://docs.keel.so/functions
const hooks: GetPersonBeforeQueryHooks = {
  beforeQuery: (ctx, inputs, query) => {
    return query.where({
      sex: Sex.Female,
    });
  },
};

export default GetPersonBeforeQuery(hooks);
