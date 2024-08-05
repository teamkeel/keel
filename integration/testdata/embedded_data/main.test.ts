import { actions, models, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

test("get - belongs to & has many", async () => {
  const post = await actions.createPost({
    title: "foo",
    content: "bcd",
    category: {
      title: "Test",
    },
    comments: [{ title: "comment1" }, { title: "comment2" }],
  });

  const fetchedPost = await actions.getPost({ id: post.id });
  expect(fetchedPost!.id).toEqual(post.id);
  expect(fetchedPost!.category!.title).toEqual("Test");
  expect(fetchedPost!.comments.length).toEqual(2);

  const comments = fetchedPost!.comments.map((x) => x.title);
  expect(comments).toHaveLength(2);
  expect(comments).toContain("comment1");
  expect(comments).toContain("comment2");
});

test("list action - belongs to", async () => {
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
  expect(posts.results[1].category!.title).toEqual("Testing again");
});

test("get - same table embed", async () => {
  const parent = await actions.createPost({
    title: "foo",
    content: "bcd",
    category: {
      title: "Test",
    },
    comments: [{ title: "comment1" }, { title: "comment2" }],
  });

  const child = await actions.createPost({
    title: "fooChild",
    content: "bcdChild",
    category: {
      title: "Test2",
    },
    comments: [],
    parent: { id: parent.id },
  });

  const fetchedPost = await actions.getPost({ id: child.id });
  expect(fetchedPost!.id).toEqual(child.id);
  expect(fetchedPost!.category!.title).toEqual("Test2");
  expect(fetchedPost!.parent!.id).toEqual(parent.id);
  expect(fetchedPost!.parent!.title).toEqual("foo");
});
