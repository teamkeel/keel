import { ListBooksBeforeQueryReturnValues } from "@teamkeel/sdk";

// This function is testing that the beforeQuery hook of a list
// function can return a list of objects, rather than a QueryBuilder instance
export default ListBooksBeforeQueryReturnValues({
  beforeQuery() {
    return [
      {
        id: "1234",
        createdAt: new Date("2001-01-01"),
        updatedAt: new Date("2001-01-01"),
        title: "Dreamcatcher",
        published: true,
      },
    ];
  },
});
