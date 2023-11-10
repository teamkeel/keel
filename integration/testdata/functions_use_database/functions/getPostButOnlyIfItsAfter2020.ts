import { GetPostButOnlyIfItsAfter2020, useDatabase } from "@teamkeel/sdk";

export default GetPostButOnlyIfItsAfter2020({
  beforeQuery: async (ctx, inputs, query) => {
    const db = useDatabase();

    // Kysely provides a CamelCasePlugin which automatically converts quueries with
    // column names written in camelCase to the underlying snake casing at database level
    const post = await db
      .selectFrom("post")
      .selectAll()
      .where("id", "=", inputs.id)
      .where("createdAt", ">=", new Date(2020, 1, 1))
      .executeTakeFirstOrThrow();

    return post;
  },
});
