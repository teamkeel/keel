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
    message: "Post field 'subTitle' can only contain unique values",
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
