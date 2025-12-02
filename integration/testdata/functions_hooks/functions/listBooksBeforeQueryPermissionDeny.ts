import { ListBooksBeforeQueryPermissionDeny, permissions } from "@teamkeel/sdk";

export default ListBooksBeforeQueryPermissionDeny({
  beforeQuery(ctx, inputs, query) {
    permissions.deny();
  },
});
