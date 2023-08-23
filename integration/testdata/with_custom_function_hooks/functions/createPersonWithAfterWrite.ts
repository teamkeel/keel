import { CreatePersonWithAfterWrite, Sex, permissions } from "@teamkeel/sdk";

export default CreatePersonWithAfterWrite({
  afterWrite: async (ctx, inputs, person) => {
    // person is retrieved from the db after insertion
    if (person.sex === Sex.Male) {
      permissions.deny();
    }
  },
});
