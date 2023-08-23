import {
  DeletePersonBeforeQuery,
  DeletePersonBeforeQueryHooks,
  Sex,
} from "@teamkeel/sdk";

// To learn more about what you can do with hooks,
// visit https://docs.keel.so/functions
const hooks: DeletePersonBeforeQueryHooks = {
  beforeQuery: (ctx, inputs, query) => {
    // mutate the query to add a constraint that doesn't match the test
    return query.where({
      sex: Sex.Female,
    });
  },
};

export default DeletePersonBeforeQuery(hooks);
