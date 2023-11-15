import { ListBooksAfterQueryPermissions, permissions } from "@teamkeel/sdk";

export default ListBooksAfterQueryPermissions({
  afterQuery(ctx, inputs, records) {
    for (const r of records) {
      if (inputs.where.onlyPublished && !r.published) {
        permissions.deny();
      }
    }

    return records;
  },
});
