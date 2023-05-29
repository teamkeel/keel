import { actions, models, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

test("permission expression with create in M:1 relationship - related model satisfies condition - authorization successful", async () => {
  const author = await models.author.create({
    name: "Keelson",
    isActive: true,
  });

  const createPost = await actions.createPost({
    title: "New Post",
    theAuthor: { value: { id: author.id } },
  });
  const collection = await models.post.findMany({});

  expect(createPost.theAuthorId).toEqual(author.id);
  expect(collection[0].id).toEqual(createPost.id);
  expect(collection[0].theAuthorId).toEqual(author.id);
});

test("permission expression with create in M:1 relationship - related model does not satisfy condition - authorization not successful", async () => {
  const author = await models.author.create({
    name: "Keelson",
    isActive: false,
  });

  await expect(
    actions.createPost({
      title: "New Post",
      theAuthor: { value: { id: author.id } },
    })
  ).toHaveAuthorizationError();

  const collection = await models.post.findMany({});
  expect(collection.length).toEqual(0);
});

test("permission expression in M:1 relationship - all related models satisfy condition - authorization successful", async () => {
  const author1 = await models.author.create({
    name: "Keelson",
    isActive: true,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    isActive: true,
  });
  const post1 = await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  const post2 = await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  const post3 = await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const { results: posts } = await actions.listPosts({});
  expect(posts.length).toEqual(3);

  const getPost1 = await actions.getPost({ id: post1.id });
  expect(getPost1!.id).toEqual(post1.id);

  const getPost2 = await actions.getPost({ id: post2.id });
  expect(getPost2!.id).toEqual(post2.id);

  const getPost3 = await actions.getPost({ id: post3.id });
  expect(getPost3!.id).toEqual(post3.id);
});

test("permission expression in M:1 relationship - Weaveton author not active - authorization response on Weaveton post", async () => {
  const author1 = await models.author.create({
    name: "Keelson",
    isActive: true,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    isActive: false,
  });
  const post1 = await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  const post2 = await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  const post3 = await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  await expect(actions.listPosts({})).toHaveAuthorizationError();

  const getPost1 = await actions.getPost({ id: post1.id });
  expect(getPost1!.id).toEqual(post1.id);

  const getPost2 = await actions.getPost({ id: post2.id });
  expect(getPost2!.id).toEqual(post2.id);

  await expect(actions.getPost({ id: post3.id })).toHaveAuthorizationError();
});

test("permission expression in M:1 relationship - posts not active - authorization successful", async () => {
  const author1 = await models.author.create({
    name: "Keelson",
    isActive: true,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    isActive: true,
  });
  const post1 = await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: false,
  });
  const post2 = await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: false,
  });
  const post3 = await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: false,
  });

  const { results: posts } = await actions.listPosts({});
  expect(posts.length).toEqual(3);

  const getPost1 = await actions.getPost({ id: post1.id });
  expect(getPost1!.id).toEqual(post1.id);

  const getPost2 = await actions.getPost({ id: post2.id });
  expect(getPost2!.id).toEqual(post2.id);

  const getPost3 = await actions.getPost({ id: post3.id });
  expect(getPost3!.id).toEqual(post3.id);
});

test("permission expression in 1:M relationship - all related models satisfy condition - authorization successful", async () => {
  const author1 = await models.author.create({
    name: "Keelson",
    isActive: true,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    isActive: true,
  });
  await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const { results: authors } = await actions.listAuthors({});
  expect(authors.length).toEqual(2);

  const getAuthor1 = await actions.getAuthor({ id: author1.id });
  expect(getAuthor1!.id).toEqual(author1.id);

  const getAuthor2 = await actions.getAuthor({ id: author2.id });
  expect(getAuthor2!.id).toEqual(author2.id);
});

test("permission expression in 1:M relationship - Weaveton post not active - authorization response on Weaveton author", async () => {
  const author1 = await models.author.create({
    name: "Keelson",
    isActive: true,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    isActive: true,
  });
  const post1 = await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  const post2 = await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  const post3 = await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: false,
  });

  await expect(actions.listAuthors({})).toHaveAuthorizationError();

  const getAuthor1 = await actions.getAuthor({ id: author1.id });
  expect(getAuthor1!.id).toEqual(author1.id);

  await expect(
    actions.getAuthor({ id: author2.id })
  ).toHaveAuthorizationError();
});

test("permission expression in 1:M relationship - one Keelson post not active - authorization successful", async () => {
  const author1 = await models.author.create({
    name: "Keelson",
    isActive: true,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    isActive: true,
  });
  const post1 = await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  const post2 = await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: false,
  });
  const post3 = await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const { results: authors } = await actions.listAuthors({});
  expect(authors.length).toEqual(2);

  const getAuthor1 = await actions.getAuthor({ id: author1.id });
  expect(getAuthor1!.id).toEqual(author1.id);

  const getAuthor2 = await actions.getAuthor({ id: author2.id });
  expect(getAuthor2!.id).toEqual(author2.id);
});

test("permission expression in 1:M relationship - all Keelsons post not active - authorization response on Keelson author", async () => {
  const author1 = await models.author.create({
    name: "Keelson",
    isActive: true,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    isActive: true,
  });
  const post1 = await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: false,
  });
  const post2 = await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: false,
  });
  const post3 = await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  await expect(actions.listAuthors({})).toHaveAuthorizationError();

  await expect(
    actions.getAuthor({ id: author1.id })
  ).toHaveAuthorizationError();

  const getAuthor2 = await actions.getAuthor({ id: author2.id });
  expect(getAuthor2!.id).toEqual(author2.id);
});
