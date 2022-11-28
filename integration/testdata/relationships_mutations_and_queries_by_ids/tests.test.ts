import {
  test,
  expect,
  actions,
  Post,
  Author,
} from "@teamkeel/testing";


test("create with parent id as implicit input - get by id - parent id set correctly", async () => {
  const { object: author } = await actions.createAuthor({ name: "Keelson" });
  const { object: post } = await actions.createPost({ title: "Keelson Post", theAuthorId: author.id });

  expect(post.theAuthorId).toEqual(author.id);

  const { object: getPost } = await actions.getPost({ id: post.id });
  expect(getPost.id).toEqual(post.id);
  expect(getPost.theAuthorId).toEqual(author.id);
  expect(getPost.title).toEqual("Keelson Post");
});

test("create with parent id as implicit input foreign key - get by id - parent id set correctly", async () => {
  const { object: author } = await actions.createAuthor({ name: "Keelson" });
  const { object: post } = await actions.createPostForeignKey({ title: "Keelson Post", theAuthorId: author.id });

  expect(post.theAuthorId).toEqual(author.id);

  const { object: getPost } = await actions.getPost({ id: post.id });
  expect(getPost.id).toEqual(post.id);
  expect(getPost.theAuthorId).toEqual(author.id);
  expect(getPost.title).toEqual("Keelson Post");
});

test("create with parent id with set attribute - get by id - parent id set correctly", async () => {
  const { object: author } = await actions.createAuthor({ name: "Keelson" });
  const { object: post } = await actions.createPostWithSet({ title: "Keelson Post", explicitAuthorId: author.id });

  expect(post.theAuthorId).toEqual(author.id);

  const { object: getPost } = await actions.getPost({ id: post.id });
  expect(getPost.id).toEqual(post.id);
  expect(getPost.theAuthorId).toEqual(author.id);
  expect(getPost.title).toEqual("Keelson Post");
});


test("update parent id as implicit input - get by id - parent id updated correctly", async () => {
  const { object: author1 } = await actions.createAuthor({ name: "Keelson" });
  const { object: post } = await actions.createPost({ title: "Keelson Post", theAuthorId: author1.id });
  const { object: author2 } = await actions.createAuthor({ name: "Weaveton" });

  const { object: getPost } = await actions.getPost({ id: post.id });
  expect(getPost.id).toEqual(post.id);
  expect(getPost.theAuthorId).toEqual(author1.id);
  expect(getPost.title).toEqual("Keelson Post");

  const { object: updatePost } = await actions.updatePost({ 
    where: { id: post.id }, 
    values: { title: "Updated", theAuthorId: author2.id } 
  });

  const { object: getUpdatedPost } = await actions.getPost({ id: post.id });
  expect(getUpdatedPost.id).toEqual(post.id);
  expect(getUpdatedPost.theAuthorId).toEqual(author2.id);
  expect(getUpdatedPost.title).toEqual("Updated");
});

test("update parent id as implicit input foreign key - get by id - parent id updated correctly", async () => {
  const { object: author1 } = await actions.createAuthor({ name: "Keelson" });
  const { object: post } = await actions.createPost({ title: "Keelson Post", theAuthorId: author1.id });
  const { object: author2 } = await actions.createAuthor({ name: "Weaveton" });

  const { object: getPost } = await actions.getPost({ id: post.id });
  expect(getPost.id).toEqual(post.id);
  expect(getPost.theAuthorId).toEqual(author1.id);
  expect(getPost.title).toEqual("Keelson Post");

  const { object: updatePost } = await actions.updatePostForeignKey({ 
    where: { id: post.id }, 
    values: { title: "Updated", theAuthorId: author2.id } 
  });

  const { object: getUpdatedPost } = await actions.getPost({ id: post.id });
  expect(getUpdatedPost.id).toEqual(post.id);
  expect(getUpdatedPost.theAuthorId).toEqual(author2.id);
  expect(getUpdatedPost.title).toEqual("Updated");
});

test("update parent id as implicit input with set attribute - get by id - parent id updated correctly", async () => {
  const { object: author1 } = await actions.createAuthor({ name: "Keelson" });
  const { object: post } = await actions.createPost({ title: "Keelson Post", theAuthorId: author1.id });
  const { object: author2 } = await actions.createAuthor({ name: "Weaveton" });

  const { object: getPost } = await actions.getPost({ id: post.id });
  expect(getPost.id).toEqual(post.id);
  expect(getPost.theAuthorId).toEqual(author1.id);
  expect(getPost.title).toEqual("Keelson Post");

  const { object: updatePost } = await actions.updatePostWithSet({ 
    where: { id: post.id }, 
    values: { title: "Updated", explicitAuthorId: author2.id } 
  });

  const { object: getUpdatedPost } = await actions.getPost({ id: post.id });
  expect(getUpdatedPost.id).toEqual(post.id);
  expect(getUpdatedPost.theAuthorId).toEqual(author2.id);
  expect(getUpdatedPost.title).toEqual("Updated");
});

test("get filter by parent id - get by id and parent id - filtered correctly", async () => {
  const { object: author1 } = await actions.createAuthor({ name: "Keelson" });
  const { object: post1 } = await actions.createPost({ title: "Keelson Post", theAuthorId: author1.id });
  const { object: author2 } = await actions.createAuthor({ name: "Weaveton" });
  const { object: post2 } = await actions.createPost({ title: "Weaveton Post", theAuthorId: author2.id });

  const { object: getPost1 } = await actions.getPostByAuthor({ id: post1.id, theAuthorId: author1.id });
  expect(getPost1.id).toEqual(post1.id);
  expect(getPost1.theAuthorId).toEqual(author1.id);

  expect(
    await actions.getPostByAuthor({ id: post1.id, theAuthorId: author2.id })
  ).toHaveError({
    message: "no records found for Get() operation",
  });
});

test("list filter by parent id - list and parent id - filtered correctly", async () => {
  const { object: author1 } = await actions.createAuthor({ name: "Keelson" });
  const { object: post1 } = await actions.createPost({ title: "Keelson Post", theAuthorId: author1.id });
  const { object: author2 } = await actions.createAuthor({ name: "Weaveton" });
  const { object: post2 } = await actions.createPost({ title: "Weaveton Post", theAuthorId: author2.id });

  const { collection: listPost } = await actions.listPost({ where: { theAuthorId: { equals: author1.id } } });
  expect(listPost.length).toEqual(1);
  expect(listPost[0].id).toEqual(post1.id);
  expect(listPost[0].theAuthorId).toEqual(author1.id);
  expect(listPost[0].title).toEqual(post1.title);
});

test("get filter by child id - get by id and parent id - filtered correctly", async () => {
  const { object: author1 } = await actions.createAuthor({ name: "Keelson" });
  const { object: post1 } = await actions.createPost({ title: "Keelson Post", theAuthorId: author1.id });
  const { object: author2 } = await actions.createAuthor({ name: "Weaveton" });
  const { object: post2 } = await actions.createPost({ title: "Weaveton Post", theAuthorId: author2.id });

  const { object: getAuthor1 } = await actions.getAuthorByPost({ id: author1.id, thePostsId: post1.id });
  expect(getAuthor1.id).toEqual(author1.id);
  expect(getAuthor1.name).toEqual(author1.name);

  expect(
    await actions.getAuthorByPost({ id: author1.id, thePostsId: post2.id })
  ).toHaveError({
    message: "no records found for Get() operation",
  });
});

test("list filter by parent id - list and parent id - filtered correctly", async () => {
  const { object: author1 } = await actions.createAuthor({ name: "Keelson" });
  const { object: post1 } = await actions.createPost({ title: "Keelson Post", theAuthorId: author1.id });
  const { object: author2 } = await actions.createAuthor({ name: "Weaveton" });
  const { object: post2 } = await actions.createPost({ title: "Weaveton Post", theAuthorId: author2.id });

  const { collection: listAuthor } = await actions.listAuthors({ where: { thePostsId: { equals: post1.id } } });
  expect(listAuthor.length).toEqual(1);
  expect(listAuthor[0].id).toEqual(author1.id);
  expect(listAuthor[0].name).toEqual(author1.name);
});