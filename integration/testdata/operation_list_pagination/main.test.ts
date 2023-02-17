import { test, expect, beforeEach } from "vitest";
import { Post } from "@teamkeel/sdk";
import { actions, resetDatabase, models } from "@teamkeel/testing";

beforeEach(resetDatabase);

async function setupPosts({ count }: { count: number }): Promise<Post[]> {
  return await Promise.all(
    Array.from(Array(count)).map(async (_, i) => {
      const p = await models.post.create({
        id: (i + 1).toString(),
        title: `Post ${i}`,
      });
      return p;
    })
  );
}

test("pagination - before", async () => {
  const posts = await setupPosts({ count: 6 });
  const { endCursor } = await actions.listPosts({
    first: 4,
  });

  const { results } = await actions.listPosts({
    before: endCursor,
  });

  expect(results.length).toEqual(3);

  expect(results.map((r) => r.id)).toEqual(posts.map((p) => p.id).slice(0, 3));
});

test("pagination - last with before", async () => {
  const posts = await setupPosts({ count: 6 });
  const { endCursor, results: firstResults } = await actions.listPosts({
    first: 4,
  });

  const { results } = await actions.listPosts({
    last: 1,
    before: endCursor,
  });

  expect(results.length).toEqual(1);

  expect(results.map((r) => r.id)).toEqual(["3"]);
});

test("pagination - first", async () => {
  const posts = await setupPosts({ count: 6 });

  const { results } = await actions.listPosts({
    first: 2,
  });

  expect(results.length).toEqual(2);
  expect(results.map((r) => r.id)).toEqual(posts.map((p) => p.id).slice(0, 2));
});

test("pagination - last only", async () => {
  const posts = await setupPosts({ count: 6 });

  const { results } = await actions.listPosts({
    last: 2,
  });

  expect(results.length).toEqual(2);
  expect(results.map((r) => r.id)).toEqual(
    posts.map((p) => p.id).slice(posts.length - 2, posts.length)
  );
});

test("pagination - first with after", async () => {
  const posts = await setupPosts({ count: 6 });
  const { endCursor } = await actions.listPosts({
    first: 4,
  });

  const { results } = await actions.listPosts({
    first: 1,
    after: endCursor,
  });

  expect(results.length).toEqual(1);
  expect(results.map((r) => r.id)).toEqual(
    posts.map((p) => p.id).slice(posts.length - 2, posts.length - 1)
  );
});

test("pagination - after", async () => {
  const posts = await setupPosts({ count: 6 });
  const { endCursor } = await actions.listPosts({
    first: 3,
  });

  const { results } = await actions.listPosts({
    after: endCursor,
  });

  expect(results.length).toEqual(3);

  expect(results.map((r) => r.id)).toEqual(
    posts.map((p) => p.id).slice(posts.length - 3, posts.length)
  );
});
