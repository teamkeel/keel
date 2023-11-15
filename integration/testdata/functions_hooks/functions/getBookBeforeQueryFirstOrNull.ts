import { GetBookBeforeQueryFirstOrNull, models } from "@teamkeel/sdk";

// This function is testing that the beforeQuery hook of a get a
// function can return null (to indicate no record found) or a record
export default GetBookBeforeQueryFirstOrNull({
  async beforeQuery(ctx, inputs) {
    const books = await models.book.findMany({
      where: {
        title: {
          contains: inputs.title,
        },
      },
    });
    if (books.length === 0) {
      return null;
    }

    return models.book.findOne({
      id: books[0].id,
    });
  },
});
