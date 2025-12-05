import { GetBookBeforeQueryPermissionDeny, permissions } from "@teamkeel/sdk";

export default GetBookBeforeQueryPermissionDeny({
  beforeQuery(ctx, inputs, query) {
    permissions.deny();
  },
});
