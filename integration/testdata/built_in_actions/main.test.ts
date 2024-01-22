import { actions, models, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";
import { Category } from "@teamkeel/sdk";

beforeEach(resetDatabase);

test("create action", async () => {
  const createdPost = await actions.createPost({
    title: "foo",
    subTitle: "abc",
  });

  expect(createdPost.title).toEqual("foo");
});

test("create action - required field is null - returns error", async () => {
  await actions.createPost({
    title: "foo",
    subTitle: "not unique",
  });

  await expect(
    actions.createPost({
      title: "foo2",
      subTitle: "not unique",
    })
  ).toHaveError({
    code: "ERR_INVALID_INPUT",
    message: "the value for the unique field 'subTitle' must be unique",
  });
});

test("get action", async () => {
  const post = await actions.createPost({
    title: "foo",
    subTitle: "bcd",
  });

  const fetchedPost = await actions.getPost({ id: post.id });
  expect(fetchedPost!.id).toEqual(post.id);
});

test("get action - no result", async () => {
  const fetchedPost = await actions.getPost({ id: "1234" });
  expect(fetchedPost).toEqual(null);
});

test("list action - equals", async () => {
  await models.post.create({ title: "apple", subTitle: "def" });
  await models.post.create({ title: "apple", subTitle: "efg" });

  const posts = await actions.listPosts({
    where: {
      title: { equals: "apple" },
    },
  });

  expect(posts.results.length).toEqual(2);
});

test("list action - notEquals on id", async () => {
  await models.post.create({ title: "apple", subTitle: "def" });
  const orange = await models.post.create({ title: "orange", subTitle: "efg" });

  const posts = await actions.listPosts({
    where: {
      id: { notEquals: orange.id },
    },
  });

  expect(posts.results.length).toEqual(1);
  expect(posts.results[0].title).toEqual("apple");
});

test("list action - notEquals with string", async () => {
  await models.post.create({ title: "apple", subTitle: "def" });
  await models.post.create({ title: "orange", subTitle: "efg" });

  const posts = await actions.listPosts({
    where: {
      title: { notEquals: "apple" },
    },
  });

  expect(posts.results.length).toEqual(1);
});

test("list action - notEquals with number", async () => {
  await models.post.create({ title: "apple", subTitle: "def", rating: 10 });
  await models.post.create({ title: "orange", subTitle: "efg", rating: 9 });

  const posts = await actions.listPosts({
    where: {
      rating: { notEquals: 10 },
    },
  });

  expect(posts.results.length).toEqual(1);
  expect(posts.results[0].title).toEqual("orange");
});

test("list action - notEquals with enum", async () => {
  await models.post.create({
    title: "pear",
    category: Category.Technical,
    subTitle: "lmn",
  });
  await models.post.create({
    title: "mango",
    category: Category.Lifestyle,
    subTitle: "mno",
  });

  const posts = await actions.listPosts({
    where: {
      category: { notEquals: Category.Lifestyle },
    },
  });

  expect(posts.results.length).toEqual(1);
  expect(posts.results[0].category).toEqual(Category.Technical);
});

test("list action - notEquals with ID", async () => {
  const otherPost = await models.post.create({
    title: "apple",
    subTitle: "def",
    rating: 10,
  });
  const post = await models.post.create({
    title: "orange",
    subTitle: "efg",
    rating: 9,
  });

  const posts = await actions.listPosts({
    where: {
      id: { notEquals: post.id },
    },
  });

  expect(posts.results.length).toEqual(1);
  expect(posts.results[0].id).toEqual(otherPost.id);
});

test("list action - contains", async () => {
  await models.post.create({ title: "banan", subTitle: "fgh" });
  await models.post.create({ title: "banana", subTitle: "ghi" });

  const { results } = await actions.listPosts({
    where: {
      title: { contains: "ana" },
    },
  });

  expect(results.length).toEqual(2);
});

test("list action - startsWith", async () => {
  await models.post.create({ title: "adam", subTitle: "hij" });
  await models.post.create({ title: "adamant", subTitle: "ijk" });
  await models.post.create({ title: "notadam", subTitle: "sed" });

  const { results } = await actions.listPosts({
    where: {
      title: { startsWith: "adam" },
    },
  });

  expect(results.length).toEqual(2);
});

test("list action - endsWith", async () => {
  await models.post.create({ title: "star wars", subTitle: "jkl" });
  await models.post.create({
    title: "a post about star wars",
    subTitle: "klm",
  });

  const { results } = await actions.listPosts({
    where: {
      title: { endsWith: "star wars" },
    },
  });

  expect(results.length).toEqual(2);
});

test("list action - oneOf text", async () => {
  await models.post.create({ title: "pear", subTitle: "lmn" });
  await models.post.create({ title: "mango", subTitle: "mno" });
  await models.post.create({ title: "orange", subTitle: "fog" });

  const { results } = await actions.listPosts({
    where: {
      title: { oneOf: ["pear", "mango"] },
    },
  });

  expect(results.length).toEqual(2);
});

test("list action - oneOf enum", async () => {
  await models.post.create({
    title: "pear",
    category: Category.Technical,
    subTitle: "lmn",
  });
  await models.post.create({
    title: "mango",
    category: Category.Lifestyle,
    subTitle: "mno",
  });
  await models.post.create({
    title: "orange",
    category: Category.Food,
    subTitle: "fog",
  });

  const { results } = await actions.listPosts({
    where: {
      category: { oneOf: [Category.Technical, Category.Lifestyle] },
    },
  });

  expect(results.length).toEqual(2);
});

test("delete action", async () => {
  const post = await models.post.create({
    title: "pear",
    subTitle: "nop",
  });

  const deletedId = await actions.deletePost({ id: post.id });

  expect(deletedId).toEqual(post.id);
});

test("delete action on other unique field", async () => {
  const post = await models.post.create({
    title: "pear",
    subTitle: "nop",
  });

  const deletedId = await actions.deletePostBySubTitle({
    subTitle: post.subTitle,
  });

  expect(deletedId).toEqual(post.id);
});

test("update action", async () => {
  const post = await models.post.create({
    title: "watermelon",
    subTitle: "opm",
  });

  const updatedPost = await actions.updatePost({
    where: { id: post.id },
    values: { title: "big watermelon" },
  });

  expect(updatedPost.id).toEqual(post.id);
  expect(updatedPost.title).toEqual("big watermelon");
  expect(updatedPost.subTitle).toEqual("opm");
});

test("update action - updatedAt set", async () => {
  const post = await models.post.create({
    title: "watermelon",
    subTitle: "opm",
  });

  expect(post.updatedAt).not.toBeNull();
  expect(post.updatedAt).toEqual(post.createdAt);

  await delay(100);

  const updatedPost = await actions.updatePost({
    where: { id: post.id },
    values: { title: "big watermelon" },
  });

  expect(updatedPost.updatedAt.valueOf()).toBeGreaterThanOrEqual(
    post.createdAt.valueOf() + 100
  );
  expect(updatedPost.updatedAt.valueOf()).toBeLessThan(
    post.createdAt.valueOf() + 1000
  );
});

test("update action - explicit set / args", async () => {
  const post = await models.post.create({
    title: "watermelon",
    subTitle: "opm",
  });

  const updatedPost = await actions.updateWithExplicitSet({
    where: { id: post.id },
    values: { coolTitle: "a really cool title" },
  });

  expect(updatedPost.title).toEqual("a really cool title");
});

function delay(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
