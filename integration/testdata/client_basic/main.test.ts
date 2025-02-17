import { actions, models, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach, beforeAll } from "vitest";
import { Category } from "@teamkeel/sdk";
import { APIClient } from "./keelClient";

var client: APIClient;

beforeEach(() => {
  client = new APIClient({ baseUrl: process.env.KEEL_TESTING_CLIENT_API_URL! });
});

beforeEach(resetDatabase);

test("client - create action", async () => {
  const post = await client.api.mutations.createPost({
    title: "My Post",
    field1: "test",
  });

  expect(post.data).not.toBeNull();
  expect(post.data?.title).toEqual("My Post");
  expect(post.data?.views).toEqual(0);
  expect(post.data?.category).toBeNull();
  expect(post.data?.field1).toEqual("test");

  const retrieved = await models.post.findOne({ id: post.data!.id });
  expect(retrieved).not.toBeNull();
  expect(retrieved?.title).toEqual("My Post");
  expect(retrieved?.views).toEqual(0);
  expect(retrieved?.category).toBeNull();
});

test("client - get action", async () => {
  const post = await client.api.mutations.createPost({ title: "My Post" });

  const retrieved = await client.api.queries.getPost({ id: post.data!.id });
  expect(retrieved.data).not.toBeNull();
  expect(retrieved.data?.title).toEqual("My Post");
  expect(retrieved.data?.views).toEqual(0);
  expect(retrieved.data?.category).toBeNull();
});

test("client - update action", async () => {
  const post = await client.api.mutations.createPost({
    title: "My Post",
    field1: "test",
  });

  const updated = await client.api.mutations.updatePost({
    where: { id: post.data!.id },
    values: {
      title: "Updated Post",
      views: 10,
      category: Category.Lifestyle,
      field1: "test again",
    },
  });

  expect(updated.data).not.toBeNull();
  expect(updated.data?.title).toEqual("Updated Post");
  expect(updated.data?.field1).toEqual("test again");
  expect(updated.data?.views).toEqual(10);
  expect(updated.data?.category).toEqual(Category.Lifestyle);

  const retrieved = await models.post.findOne({ id: post.data!.id });
  expect(retrieved).not.toBeNull();
  expect(retrieved?.title).toEqual("Updated Post");
  expect(retrieved?.views).toEqual(10);
  expect(retrieved?.category).toEqual(Category.Lifestyle);
});

test("client - delete action", async () => {
  const post = await client.api.mutations.createPost({ title: "My Post" });

  const deleted = await client.api.mutations.deletePost({ id: post.data!.id });

  expect(deleted.data).toEqual(post.data?.id);

  const retrieved = await models.post.findOne({ id: post.data!.id });
  expect(retrieved).toBeNull();
});

test("client - list action", async () => {
  for (let i = 0; i < 101; i++) {
    await client.api.mutations.createPost({
      title: "Post " + i,
      category: Category.Food,
      views: i,
    });
  }

  const result = await client.api.queries.listPosts({
    where: { title: { startsWith: "Post" } },
  });
  expect(result.data?.results).toHaveLength(50);
  expect(result.data?.pageInfo.count).toEqual(50);
  expect(result.data?.pageInfo.totalCount).toEqual(101);
  expect(result.data?.pageInfo.hasNextPage).toBeTruthy();
  expect(result.data?.pageInfo.startCursor).not.toBeNull();
  expect(result.data?.pageInfo.endCursor).not.toBeNull();

  expect(result.data?.resultInfo.category).toEqual([
    { value: "Food", count: 101 },
  ]);

  expect(result.data?.resultInfo.views.min).toEqual(0);
  expect(result.data?.resultInfo.views.max).toEqual(100);
  expect(result.data?.resultInfo.views.avg).toEqual(50);
});

test("client - list action with paging", async () => {
  for (let i = 0; i < 101; i++) {
    await client.api.mutations.createPost({ title: "Post " + i });
  }

  const page1 = await client.api.queries.listPosts({
    where: { title: { startsWith: "Post" } },
  });
  expect(page1.data?.results).toHaveLength(50);
  expect(page1.data?.pageInfo.count).toEqual(50);
  expect(page1.data?.pageInfo.totalCount).toEqual(101);
  expect(page1.data?.pageInfo.hasNextPage).toBeTruthy();

  const page2 = await client.api.queries.listPosts({
    where: { title: { startsWith: "Post" } },
    after: page1.data?.pageInfo.endCursor,
  });
  expect(page2.data?.results).toHaveLength(50);
  expect(page2.data?.pageInfo.count).toEqual(50);
  expect(page2.data?.pageInfo.totalCount).toEqual(101);
  expect(page2.data?.pageInfo.hasNextPage).toBeTruthy();

  const page3 = await client.api.queries.listPosts({
    where: { title: { startsWith: "Post" } },
    after: page2.data?.pageInfo.endCursor,
  });
  expect(page3.data?.results).toHaveLength(1);
  expect(page3.data?.pageInfo.count).toEqual(1);
  expect(page3.data?.pageInfo.totalCount).toEqual(101);
  expect(page3.data?.pageInfo.hasNextPage).toBeFalsy();
});

test("client - list action large page", async () => {
  for (let i = 0; i < 101; i++) {
    await client.api.mutations.createPost({ title: "Post " + i });
  }

  const result = await client.api.queries.listPosts({
    where: { title: { startsWith: "Post" } },
    first: 200,
  });
  expect(result.data?.results).toHaveLength(101);
  expect(result.data?.pageInfo.count).toEqual(101);
  expect(result.data?.pageInfo.totalCount).toEqual(101);
  expect(result.data?.pageInfo.hasNextPage).toBeFalsy();
  expect(result.data?.pageInfo.startCursor).not.toBeNull();
  expect(result.data?.pageInfo.endCursor).not.toBeNull();
});
