import { test, expect, beforeEach } from "vitest";
import { actions, resetDatabase, models } from "@teamkeel/testing";

beforeEach(resetDatabase);

test("pagination simple", async () => {
  const posts = await Promise.all(
    Array.from(Array(20)).map((_, i) =>
      models.post.create({ title: `Post ${i}` })
    )
  );

  const { results } = await actions.listPosts({
    after: posts[10].id,
  });

  expect(results.length).toEqual(10);
});
