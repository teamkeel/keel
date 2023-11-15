import { ListBooksAfterQuery } from "@teamkeel/sdk";

export default ListBooksAfterQuery({
  afterQuery(ctx, inputs, records) {
    return records.map((r) => {
      return {
        ...r,
        title: r.title.toUpperCase(),
      };
    });
  },
});
