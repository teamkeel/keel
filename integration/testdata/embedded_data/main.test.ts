import { actions, models, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

test("get action with embedded data", async () => {
  const post = await actions.createPost({
    title: "foo",
    content: "bcd",
    category: {
      title: "Test",
    },
  });

  const fetchedPost = await actions.getPost({ id: post.id });
  expect(fetchedPost!.id).toEqual(post.id);
  expect(fetchedPost!.categoryId).toBeUndefined();
  expect(fetchedPost!.category!.title).toEqual("Test");
});

test("list action with embedded data", async () => {
  const post = await actions.createPost({
    title: "foo",
    content: "bcd",
    category: {
      title: "Test",
    },
  });

  const fetchedPost = await actions.getPost({ id: post.id });
  expect(fetchedPost!.id).toEqual(post.id);
  expect(fetchedPost!.categoryId).toBeUndefined();
  expect(fetchedPost!.category!.title).toEqual("Test");
});

test("list action - equals", async () => {
  await models.post.create({
    title: "foo",
    content: "bcd",
    category: {
      title: "Test",
    },
    order: 1,
  });
  await models.post.create({
    title: "Another test",
    content: "content",
    category: {
      title: "Testing again",
    },
    order: 2,
  });

  const posts = await actions.listPosts({
    orderBy: [{ order: "asc" }],
  });

  expect(posts.results.length).toEqual(2);
  expect(posts.results[0].category!.title).toEqual("Test");
  expect(posts.results[0].categoryId).toBeUndefined();
  expect(posts.results[1].category!.title).toEqual("Testing again");
  expect(posts.results[1].categoryId).toBeUndefined();
});
