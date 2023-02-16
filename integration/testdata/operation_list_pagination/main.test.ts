import { test, expect, beforeEach } from "vitest";
import { Post } from "@teamkeel/sdk";
import { actions, resetDatabase, models } from "@teamkeel/testing";

beforeEach(resetDatabase);

test("pagination - before", async () => {
  const posts: Post[] = [];
  posts.push(await models.post.create({ id: "1", title: "Post 1" }));
  posts.push(await models.post.create({ id: "2", title: "Post 2" }));
  posts.push(await models.post.create({ id: "3", title: "Post 3" }));
  posts.push(await models.post.create({ id: "4", title: "Post 4" }));
  posts.push(await models.post.create({ id: "5", title: "Post 5" }));
  posts.push(await models.post.create({ id: "6", title: "Post 6" }));

  const cursor = posts[3].id;

  const { results } = await actions.listPosts({
    before: cursor,
  });

  expect(results.length).toEqual(3);
});

test("pagination - first", async () => {
  const posts: Post[] = [];
  posts.push(await models.post.create({ id: "1", title: "Post 1" }));
  posts.push(await models.post.create({ id: "2", title: "Post 2" }));
  posts.push(await models.post.create({ id: "3", title: "Post 3" }));
  posts.push(await models.post.create({ id: "4", title: "Post 4" }));
  posts.push(await models.post.create({ id: "5", title: "Post 5" }));
  posts.push(await models.post.create({ id: "6", title: "Post 6" }));

  const { results } = await actions.listPosts({
    first: 2,
  });

  expect(results.length).toEqual(2);
  expect(results.map((r) => r.id)).toEqual(posts.map((p) => p.id).slice(0, 2));
});

test("pagination - last", async () => {
  const posts: Post[] = [];
  posts.push(await models.post.create({ id: "1", title: "Post 1" }));
  posts.push(await models.post.create({ id: "2", title: "Post 2" }));
  posts.push(await models.post.create({ id: "3", title: "Post 3" }));
  posts.push(await models.post.create({ id: "4", title: "Post 4" }));
  posts.push(await models.post.create({ id: "5", title: "Post 5" }));
  posts.push(await models.post.create({ id: "6", title: "Post 6" }));

  const { results } = await actions.listPosts({
    last: 2,
    after: posts[posts.length - 3].id,
  });

  expect(results.length).toEqual(2);
  expect(results.map((r) => r.id)).toEqual(
    posts.map((p) => p.id).slice(posts.length - 2, posts.length)
  );
});

test("pagination - after", async () => {
  const posts: Post[] = [];
  posts.push(await models.post.create({ id: "1", title: "Post 1" }));
  posts.push(await models.post.create({ id: "2", title: "Post 2" }));
  posts.push(await models.post.create({ id: "3", title: "Post 3" }));
  posts.push(await models.post.create({ id: "4", title: "Post 4" }));
  posts.push(await models.post.create({ id: "5", title: "Post 5" }));
  posts.push(await models.post.create({ id: "6", title: "Post 6" }));

  const cursor = posts[2].id;

  const { results } = await actions.listPosts({
    after: cursor,
  });

  expect(results.length).toEqual(3);
});
