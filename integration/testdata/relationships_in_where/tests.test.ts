import { actions, models, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

// M:1 with AND conditions

test("where expression in M:1 relationship - all related models satisfy condition - all author and posts returned", async () => {
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

  const { results: posts } = await actions.listPosts();
  expect(posts.length).toEqual(3);

  const getPost1 = await actions.getPost({ id: post1.id });
  expect(getPost1!.id).toEqual(post1.id);

  const getPost2 = await actions.getPost({ id: post2.id });
  expect(getPost2!.id).toEqual(post2.id);

  const getPost3 = await actions.getPost({ id: post3.id });
  expect(getPost3!.id).toEqual(post3.id);
});

test("where expression in M:1 relationship - Weaveton author not active - weaveton author and posts not returned", async () => {
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

  const { results: posts } = await actions.listPosts();
  expect(posts.length).toEqual(2);

  const getPost1 = await actions.getPost({ id: post1.id });
  expect(getPost1!.id).toEqual(post1.id);

  const getPost2 = await actions.getPost({ id: post2.id });
  expect(getPost2!.id).toEqual(post2.id);

  const getPost3 = await actions.getPost({ id: post3.id });
  expect(getPost3).toBeNull();
});

test("where expression in M:1 relationship - posts not active - nothing returned", async () => {
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

  const { results: posts } = await actions.listPosts();
  expect(posts.length).toEqual(0);

  const getPost1 = await actions.getPost({ id: post1.id });
  expect(getPost1).toBeNull();

  const getPost2 = await actions.getPost({ id: post2.id });
  expect(getPost2).toBeNull();

  const getPost3 = await actions.getPost({ id: post3.id });
  expect(getPost3).toBeNull();
});

// M:1 with OR conditions

test("where expression in M:1 relationship with ORs - all related models satisfy condition - everything returned", async () => {
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

  const { results: posts } = await actions.listPostsORed();
  expect(posts.length).toEqual(3);

  const getPost1 = await actions.getPostORed({ id: post1.id });
  expect(getPost1!.id).toEqual(post1.id);

  const getPost2 = await actions.getPostORed({ id: post2.id });
  expect(getPost2!.id).toEqual(post2.id);

  const getPost3 = await actions.getPostORed({ id: post3.id });
  expect(getPost3!.id).toEqual(post3.id);
});

test("where expression in M:1 relationship with ORs - Weaveton author not active - everything returned", async () => {
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

  const { results: posts } = await actions.listPostsORed();
  expect(posts.length).toEqual(3);

  const getPost1 = await actions.getPostORed({ id: post1.id });
  expect(getPost1!.id).toEqual(post1.id);

  const getPost2 = await actions.getPostORed({ id: post2.id });
  expect(getPost2!.id).toEqual(post2.id);

  const getPost3 = await actions.getPostORed({ id: post3.id });
  expect(getPost3!.id).toEqual(post3.id);
});

test("where expression in M:1 relationship with ORs - posts not active - everything returned", async () => {
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

  const { results: posts } = await actions.listPostsORed();
  expect(posts.length).toEqual(3);

  const getPost1 = await actions.getPostORed({ id: post1.id });
  expect(getPost1!.id).toEqual(post1.id);

  const getPost2 = await actions.getPostORed({ id: post2.id });
  expect(getPost2!.id).toEqual(post2.id);

  const getPost3 = await actions.getPostORed({ id: post3.id });
  expect(getPost3!.id).toEqual(post3.id);
});

test("where expression in M:1 relationship with ORs - weave posts and author not active - weave author and posts not returned", async () => {
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
    isActive: false,
  });

  const { results: posts } = await actions.listPostsORed();
  expect(posts.length).toEqual(2);

  const getPost1 = await actions.getPostORed({ id: post1.id });
  expect(getPost1!.id).toEqual(post1.id);

  const getPost2 = await actions.getPostORed({ id: post2.id });
  expect(getPost2!.id).toEqual(post2.id);

  const getPost3 = await actions.getPostORed({ id: post3.id });
  expect(getPost3).toBeNull();
});

// 1:M with AND conditions

test("where expression in 1:M relationship - all related models satisfy condition - everything returned", async () => {
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

  const { results: authors } = await actions.listAuthors();
  expect(authors.length).toEqual(2);

  const getAuthor1 = await actions.getAuthor({ id: author1.id });
  expect(getAuthor1!.id).toEqual(author1.id);

  const getAuthor2 = await actions.getAuthor({ id: author2.id });
  expect(getAuthor2!.id).toEqual(author2.id);
});

test("where expression in 1:M relationship - Weaveton post not active - weaveton post and author not returned", async () => {
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

  const { results: authors } = await actions.listAuthors();
  expect(authors.length).toEqual(1);

  const getAuthor1 = await actions.getAuthor({ id: author1.id });
  expect(getAuthor1!.id).toEqual(author1.id);

  const getAuthor2 = await actions.getAuthor({ id: author2.id });
  expect(getAuthor2).toBeNull();
});

test("where expression in 1:M relationship - one Keelson post not active - everything returned", async () => {
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

  const { results: authors } = await actions.listAuthors();
  expect(authors.length).toEqual(2);

  const getAuthor1 = await actions.getAuthor({ id: author1.id });
  expect(getAuthor1!.id).toEqual(author1.id);

  const getAuthor2 = await actions.getAuthor({ id: author2.id });
  expect(getAuthor2!.id).toEqual(author2.id);
});

test("where expression in 1:M relationship - all Keelsons post not active - Keelson author and posts not returned", async () => {
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

  const { results: authors } = await actions.listAuthors();
  expect(authors.length).toEqual(1);

  const getAuthor1 = await actions.getAuthor({ id: author1.id });
  expect(getAuthor1).toBeNull();

  const getAuthor2 = await actions.getAuthor({ id: author2.id });
  expect(getAuthor2!.id).toEqual(author2.id);
});

// 1:M with OR conditions

test("where expression in 1:M relationship with ORs - all related models satisfy condition - everything returned", async () => {
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

  const { results: authors } = await actions.listAuthorsORed();
  expect(authors.length).toEqual(2);

  const getAuthor1 = await actions.getAuthorORed({ id: author1.id });
  expect(getAuthor1!.id).toEqual(author1.id);

  const getAuthor2 = await actions.getAuthorORed({ id: author2.id });
  expect(getAuthor2!.id).toEqual(author2.id);
});

test("where expression in 1:M relationship with ORs - Weaveton post not active - everything returned", async () => {
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

  const { results: authors } = await actions.listAuthorsORed();
  expect(authors.length).toEqual(2);

  const getAuthor1 = await actions.getAuthorORed({ id: author1.id });
  expect(getAuthor1!.id).toEqual(author1.id);

  const getAuthor2 = await actions.getAuthorORed({ id: author2.id });
  expect(getAuthor2!.id).toEqual(author2.id);
});

test("where expression in 1:M relationship with ORs - one Keelson post not active - everything returned", async () => {
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

  const { results: authors } = await actions.listAuthorsORed();
  expect(authors.length).toEqual(2);

  const getAuthor1 = await actions.getAuthorORed({ id: author1.id });
  expect(getAuthor1!.id).toEqual(author1.id);

  const getAuthor2 = await actions.getAuthorORed({ id: author2.id });
  expect(getAuthor2!.id).toEqual(author2.id);
});

test("where expression in 1:M relationship with ORs - all posts not active - everything returned", async () => {
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

  const { results: authors } = await actions.listAuthorsORed();
  expect(authors.length).toEqual(2);

  const getAuthor1 = await actions.getAuthorORed({ id: author1.id });
  expect(getAuthor1!.id).toEqual(author1.id);

  const getAuthor2 = await actions.getAuthorORed({ id: author2.id });
  expect(getAuthor2!.id).toEqual(author2.id);
});

test("where expression in 1:M relationship with ORs - Keelson author and Keelson posts not active - Keelson author and posts not returned", async () => {
  const author1 = await models.author.create({
    name: "Keelson",
    isActive: false,
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

  const { results: authors } = await actions.listAuthorsORed();
  expect(authors.length).toEqual(1);

  const getAuthor1 = await actions.getAuthorORed({ id: author1.id });
  expect(getAuthor1).toBeNull();

  const getAuthor2 = await actions.getAuthorORed({ id: author2.id });
  expect(getAuthor2!.id).toEqual(author2.id);
});

test("where expression in 1:M relationship with ORs - no Keelson posts, everything else satisfied expression - everything returned", async () => {
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

  const { results: authors } = await actions.listAuthorsORed();
  expect(authors.length).toEqual(2);

  const getAuthor1 = await actions.getAuthorORed({ id: author1.id });
  expect(getAuthor1!.id).toEqual(author1.id);

  const getAuthor2 = await actions.getAuthorORed({ id: author2.id });
  expect(getAuthor2!.id).toEqual(author2.id);
});
