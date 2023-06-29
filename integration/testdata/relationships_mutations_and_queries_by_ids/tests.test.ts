import { actions, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

test("create with parent id as implicit input - get by id - parent id set correctly", async () => {
  const author = await actions.createAuthor({ name: "Keelson" });
  const post = await actions.createPost({
    title: "Keelson Post",
    theAuthor: { id: author.id },
  });

  expect(post.theAuthorId).toEqual(author.id);

  const getPost = await actions.getPost({ id: post.id });
  expect(getPost!.id).toEqual(post.id);
  expect(getPost!.theAuthorId).toEqual(author.id);
  expect(getPost!.title).toEqual("Keelson Post");
});

test("create with parent id as implicit input - id does not exist - returns error", async () => {
  await expect(
    actions.createPost({
      title: "Keelson Post",
      theAuthor: { id: "2L2ar5NCPvTTEdiDYqgcpF3f5QN1" },
    })
  ).toHaveError({
    code: "ERR_INVALID_INPUT",
    message: "the relationship lookup for field 'theAuthorId' does not exist",
  });
});

test("create with parent id with set attribute - get by id - parent id set correctly", async () => {
  const author = await actions.createAuthor({ name: "Keelson" });
  const post = await actions.createPostWithSet({
    title: "Keelson Post",
    explicitAuthorId: author.id,
  });

  expect(post.theAuthorId).toEqual(author.id);

  const getPost = await actions.getPost({ id: post.id });
  expect(getPost!.id).toEqual(post.id);
  expect(getPost!.theAuthorId).toEqual(author.id);
  expect(getPost!.title).toEqual("Keelson Post");
});

test("update parent id as implicit input - get by id - parent id updated correctly", async () => {
  const author1 = await actions.createAuthor({ name: "Keelson" });
  const post = await actions.createPost({
    title: "Keelson Post",
    theAuthor: { id: author1.id },
  });
  const author2 = await actions.createAuthor({ name: "Weaveton" });

  const getPost = await actions.getPost({ id: post.id });
  expect(getPost!.id).toEqual(post.id);
  expect(getPost!.theAuthorId).toEqual(author1.id);
  expect(getPost!.title).toEqual("Keelson Post");

  const updatePost = await actions.updatePost({
    where: { id: post.id },
    values: { title: "Updated", theAuthor: { id: author2.id } },
  });

  const getUpdatedPost = await actions.getPost({ id: post.id });
  expect(getUpdatedPost!.id).toEqual(post.id);
  expect(getUpdatedPost!.theAuthorId).toEqual(author2.id);
  expect(getUpdatedPost!.title).toEqual("Updated");
});

test("update parent id as implicit input with set attribute - get by id - parent id updated correctly", async () => {
  const author1 = await actions.createAuthor({ name: "Keelson" });
  const post = await actions.createPost({
    title: "Keelson Post",
    theAuthor: { id: author1.id },
  });
  const author2 = await actions.createAuthor({ name: "Weaveton" });

  const getPost = await actions.getPost({ id: post.id });
  expect(getPost!.id).toEqual(post.id);
  expect(getPost!.theAuthorId).toEqual(author1.id);
  expect(getPost!.title).toEqual("Keelson Post");

  const updatePost = await actions.updatePostWithSet({
    where: { id: post.id },
    values: { title: "Updated", explicitAuthorId: author2.id },
  });

  const getUpdatedPost = await actions.getPost({ id: post.id });
  expect(getUpdatedPost!.id).toEqual(post.id);
  expect(getUpdatedPost!.theAuthorId).toEqual(author2.id);
  expect(getUpdatedPost!.title).toEqual("Updated");
});

test("get filter by parent id - get by id and parent id - filtered correctly", async () => {
  const author1 = await actions.createAuthor({ name: "Keelson" });
  const post1 = await actions.createPost({
    title: "Keelson Post",
    theAuthor: { id: author1.id },
  });
  const author2 = await actions.createAuthor({ name: "Weaveton" });
  const post2 = await actions.createPost({
    title: "Weaveton Post",
    theAuthor: { id: author2.id },
  });

  const getPost1 = await actions.getPostByAuthor({
    id: post1.id,
    theAuthorId: author1.id,
  });
  expect(getPost1!.id).toEqual(post1.id);
  expect(getPost1!.theAuthorId).toEqual(author1.id);

  expect(
    await actions.getPostByAuthor({ id: post1.id, theAuthorId: author2.id })
  ).toEqual(null);
});

test("list filter by parent id - list and parent id - filtered correctly", async () => {
  const author1 = await actions.createAuthor({ name: "Keelson" });
  const post1 = await actions.createPost({
    title: "Keelson Post",
    theAuthor: { id: author1.id },
  });
  const author2 = await actions.createAuthor({ name: "Weaveton" });
  const post2 = await actions.createPost({
    title: "Weaveton Post",
    theAuthor: { id: author2.id },
  });

  const { results: listPost } = await actions.listPost({
    where: { theAuthor: { id: { equals: author1.id } } },
  });
  expect(listPost.length).toEqual(1);
  expect(listPost[0].id).toEqual(post1.id);
  expect(listPost[0].theAuthorId).toEqual(author1.id);
  expect(listPost[0].title).toEqual(post1.title);
});

test("get filter by child id - get by id and parent id - filtered correctly", async () => {
  const author1 = await actions.createAuthor({ name: "Keelson" });
  const post1 = await actions.createPost({
    title: "Keelson Post",
    theAuthor: { id: author1.id },
  });
  const author2 = await actions.createAuthor({ name: "Weaveton" });
  const post2 = await actions.createPost({
    title: "Weaveton Post",
    theAuthor: { id: author2.id },
  });

  const getAuthor1 = await actions.getAuthorByPost({
    id: author1.id,
    thePostsId: post1.id,
  });
  expect(getAuthor1!.id).toEqual(author1.id);
  expect(getAuthor1!.name).toEqual(author1.name);

  expect(
    await actions.getAuthorByPost({ id: author1.id, thePostsId: post2.id })
  ).toEqual(null);
});

test("list filter by parent id - list and parent id - filtered correctly", async () => {
  const author1 = await actions.createAuthor({ name: "Keelson" });
  const post1 = await actions.createPost({
    title: "Keelson Post",
    theAuthor: { id: author1.id },
  });
  const author2 = await actions.createAuthor({ name: "Weaveton" });
  const post2 = await actions.createPost({
    title: "Weaveton Post",
    theAuthor: { id: author2.id },
  });

  const { results: listAuthor } = await actions.listAuthors({
    where: { thePosts: { id: { equals: post1.id } } },
  });
  expect(listAuthor.length).toEqual(1);
  expect(listAuthor[0].id).toEqual(author1.id);
  expect(listAuthor[0].name).toEqual(author1.name);
});
