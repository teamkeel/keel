import { GetBookAfterQueryPermissions, permissions } from "@teamkeel/sdk";

// This function is testing that permission can be denied in the afterQuery hook of a get function
export default GetBookAfterQueryPermissions({
  afterQuery(ctx, inputs, record) {
    if (inputs.onlyPublished && !record.published) {
      permissions.deny();
    }

    return record;
  },
});
