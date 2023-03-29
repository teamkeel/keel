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
  const {
    pageInfo: { endCursor },
  } = await actions.listPosts({
    first: 4,
  });

  const { results } = await actions.listPosts({
    before: endCursor,
  });

  expect(results.length).toEqual(3);

  expect(results.map((r) => r.id)).toEqual(posts.map((p) => p.id).slice(0, 3));
});

test("pagination - last with before", async () => {
  await setupPosts({ count: 6 });
  const {
    pageInfo: { endCursor },
  } = await actions.listPosts({
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

// todo: hasNextPage doesnt seem to return the correct value here
// https://github.com/teamkeel/keel/blob/055ec3629bc7e1cfb5f6284d6019cef116ac9a92/runtime/actions/query.go#L307-L308
test.fails("hasNextPage", async () => {
  await setupPosts({ count: 6 });

  const {
    pageInfo: { hasNextPage },
  } = await actions.listPosts({
    last: 2,
  });

  expect(hasNextPage).toEqual(false);
});
