import {
  UpdatePersonWithAfterQuery,
  UpdatePersonWithAfterQueryHooks,
  Sex,
} from "@teamkeel/sdk";

// To learn more about what you can do with hooks,
// visit https://docs.keel.so/functions
const hooks: UpdatePersonWithAfterQueryHooks = {
  afterQuery: async (ctx, inputs, person) => {
    return {
      ...person,
      title: "not " + person.title,
    };
  },
};

export default UpdatePersonWithAfterQuery(hooks);
