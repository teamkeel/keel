import {
  ListPeopleBeforeQuery,
  ListPeopleBeforeQueryHooks,
  Sex,
} from "@teamkeel/sdk";

// To learn more about what you can do with hooks,
// visit https://docs.keel.so/functions
const hooks: ListPeopleBeforeQueryHooks = {
  beforeQuery: (ctx, inputs, query) => {
    return query.where({
      sex: Sex.Female,
    });
  },
};

export default ListPeopleBeforeQuery(hooks);
