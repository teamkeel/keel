import { UpdatePersonWithAfterWrite, Sex, permissions } from "@teamkeel/sdk";

export default UpdatePersonWithAfterWrite({
  afterWrite: async (ctx, inputs, person) => {
    // person is retrieved from the db after insertion
    if (person.sex === Sex.Male) {
      permissions.deny();
    }
  },
});
