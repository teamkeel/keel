import { test, expect, actions, Post, Author } from "@teamkeel/testing";

test("permission expression in M:1 relationship - all related models satisfy condition - authorization successful", async () => {
  const { object: author1 } = await Author.create({
    name: "Keelson",
    isActive: true,
  });
  const { object: author2 } = await Author.create({
    name: "Weaveton",
    isActive: true,
  });
  const { object: post1 } = await Post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  const { object: post2 } = await Post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  const { object: post3 } = await Post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const { collection: posts } = await actions.listPosts({});
  expect(posts.length).toEqual(3);

  const { object: getPost1 } = await actions.getPost({ id: post1.id });
  expect(getPost1.id).toEqual(post1.id);

  const { object: getPost2 } = await actions.getPost({ id: post2.id });
  expect(getPost2.id).toEqual(post2.id);

  const { object: getPost3 } = await actions.getPost({ id: post3.id });
  expect(getPost3.id).toEqual(post3.id);
});

test("permission expression in M:1 relationship - Weaveton author not active - authorization response on Weaveton post", async () => {
  const { object: author1 } = await Author.create({
    name: "Keelson",
    isActive: true,
  });
  const { object: author2 } = await Author.create({
    name: "Weaveton",
    isActive: false,
  });
  const { object: post1 } = await Post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  const { object: post2 } = await Post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  const { object: post3 } = await Post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  expect(await actions.listPosts({})).toHaveAuthorizationError();

  const { object: getPost1 } = await actions.getPost({ id: post1.id });
  expect(getPost1.id).toEqual(post1.id);

  const { object: getPost2 } = await actions.getPost({ id: post2.id });
  expect(getPost2.id).toEqual(post2.id);

  expect(await actions.getPost({ id: post3.id })).toHaveAuthorizationError();
});

test("permission expression in M:1 relationship - posts not active - authorization successful", async () => {
  const { object: author1 } = await Author.create({
    name: "Keelson",
    isActive: true,
  });
  const { object: author2 } = await Author.create({
    name: "Weaveton",
    isActive: true,
  });
  const { object: post1 } = await Post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: false,
  });
  const { object: post2 } = await Post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: false,
  });
  const { object: post3 } = await Post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: false,
  });

  const { collection: posts } = await actions.listPosts({});
  expect(posts.length).toEqual(3);

  const { object: getPost1 } = await actions.getPost({ id: post1.id });
  expect(getPost1.id).toEqual(post1.id);

  const { object: getPost2 } = await actions.getPost({ id: post2.id });
  expect(getPost2.id).toEqual(post2.id);

  const { object: getPost3 } = await actions.getPost({ id: post3.id });
  expect(getPost3.id).toEqual(post3.id);
});

test("permission expression in 1:M relationship - all related models satisfy condition - authorization successful", async () => {
  const { object: author1 } = await Author.create({
    name: "Keelson",
    isActive: true,
  });
  const { object: author2 } = await Author.create({
    name: "Weaveton",
    isActive: true,
  });
  const { object: post1 } = await Post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  const { object: post2 } = await Post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  const { object: post3 } = await Post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const { collection: authors } = await actions.listAuthors({});
  expect(authors.length).toEqual(2);

  const { object: getAuthor1 } = await actions.getAuthor({ id: author1.id });
  expect(getAuthor1.id).toEqual(author1.id);

  const { object: getAuthor2 } = await actions.getAuthor({ id: author2.id });
  expect(getAuthor2.id).toEqual(author2.id);
});

test("permission expression in 1:M relationship - Weaveton post not active - authorization response on Weaveton author", async () => {
  const { object: author1 } = await Author.create({
    name: "Keelson",
    isActive: true,
  });
  const { object: author2 } = await Author.create({
    name: "Weaveton",
    isActive: true,
  });
  const { object: post1 } = await Post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  const { object: post2 } = await Post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  const { object: post3 } = await Post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: false,
  });

  expect(await actions.listAuthors({})).toHaveAuthorizationError();

  const { object: getAuthor1 } = await actions.getAuthor({ id: author1.id });
  expect(getAuthor1.id).toEqual(author1.id);

  expect(
    await actions.getAuthor({ id: author2.id })
  ).toHaveAuthorizationError();
});

test("permission expression in 1:M relationship - one Keelson post not active - authorization successful", async () => {
  const { object: author1 } = await Author.create({
    name: "Keelson",
    isActive: true,
  });
  const { object: author2 } = await Author.create({
    name: "Weaveton",
    isActive: true,
  });
  const { object: post1 } = await Post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  const { object: post2 } = await Post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: false,
  });
  const { object: post3 } = await Post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const { collection: authors } = await actions.listAuthors({});
  expect(authors.length).toEqual(2);

  const { object: getAuthor1 } = await actions.getAuthor({ id: author1.id });
  expect(getAuthor1.id).toEqual(author1.id);

  const { object: getAuthor2 } = await actions.getAuthor({ id: author2.id });
  expect(getAuthor2.id).toEqual(author2.id);
});

test("permission expression in 1:M relationship - all Keelsons post not active - authorization response on Keelson author", async () => {
  const { object: author1 } = await Author.create({
    name: "Keelson",
    isActive: true,
  });
  const { object: author2 } = await Author.create({
    name: "Weaveton",
    isActive: true,
  });
  const { object: post1 } = await Post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: false,
  });
  const { object: post2 } = await Post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: false,
  });
  const { object: post3 } = await Post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  expect(await actions.listAuthors({})).toHaveAuthorizationError();

  expect(
    await actions.getAuthor({ id: author1.id })
  ).toHaveAuthorizationError();

  const { object: getAuthor2 } = await actions.getAuthor({ id: author2.id });
  expect(getAuthor2.id).toEqual(author2.id);
});
