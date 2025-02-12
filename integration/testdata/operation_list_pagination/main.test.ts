import { test, expect, beforeEach } from "vitest";
import { Post } from "@teamkeel/sdk";
import { actions, resetDatabase, models } from "@teamkeel/testing";

beforeEach(resetDatabase);

async function setupPosts({ count }: { count: number }): Promise<Post[]> {
  return await Promise.all(
    Array.from(Array(count)).map(async (_, i) => {
      const p = await models.post.create({
        id: (i + 1).toString(),
        title: `Post ${i + 1}`,
      });
      return p;
    })
  );
}

test("pagination - first", async () => {
  const posts = await setupPosts({ count: 6 });

  const { results } = await actions.listPosts({
    first: 2,
  });

  expect(results.length).toEqual(2);
  expect(results.map((r) => r.id)).toEqual(posts.map((p) => p.id).slice(0, 2));
});

test("pagination - with limit", async () => {
  const posts = await setupPosts({ count: 6 });

  const { results, pageInfo } = await actions.listPosts({
    limit: 2,
  });

  expect(results.length).toEqual(2);
  expect(results.map((r) => r.id)).toEqual(posts.map((p) => p.id).slice(0, 2));
  expect(pageInfo.pageNumber).toEqual(1);
  expect(pageInfo.hasNextPage).toEqual(true);
});

test("pagination - with limit and offset", async () => {
  const posts = await setupPosts({ count: 6 });

  const { results, pageInfo } = await actions.listPosts({
    limit: 2,
    offset: 4,
  });

  expect(results.length).toEqual(2);
  expect(results.map((r) => r.id)).toEqual(posts.map((p) => p.id).slice(4, 6));
  expect(pageInfo.pageNumber).toEqual(3);
  expect(pageInfo.hasNextPage).toEqual(false);
});


test("pagination - first with after", async () => {
  const posts = await setupPosts({ count: 6 });
  const {
    pageInfo: { endCursor },
  } = await actions.listPosts({
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
  const {
    pageInfo: { endCursor, hasNextPage },
  } = await actions.listPosts({
    first: 3,
  });

  expect(endCursor).toEqual("3");
  expect(hasNextPage).toEqual(true);

  const { results } = await actions.listPosts({
    after: endCursor,
  });

  expect(results.length).toEqual(3);

  expect(results.map((r) => r.id)).toEqual(
    posts.map((p) => p.id).slice(posts.length - 3, posts.length)
  );
});

test("counts", async () => {
  const posts = await setupPosts({ count: 6 });

  const take = 3;

  const {
    pageInfo: { totalCount, count },
  } = await actions.listPosts({
    first: take,
  });

  expect(totalCount).toEqual(posts.length);
  expect(count).toEqual(take);
});

test("hasNextPage", async () => {
  await setupPosts({ count: 6 });

  const {
    pageInfo: { hasNextPage },
  } = await actions.listPosts({
    first: 2,
  });

  expect(hasNextPage).toEqual(true);
});
