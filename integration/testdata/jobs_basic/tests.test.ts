import { models, jobs, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";
import { Status } from "@teamkeel/sdk";

beforeEach(resetDatabase);

test("jobs - updating value with input - value updated", async () => {
  const post = await models.post.create({ title: "My Post" });
  await models.postViews.create({ postId: post.id, views: 3 });
  await models.postViews.create({ postId: post.id, views: 6 });
  await models.postViews.create({ postId: post.id, views: 1 });
  expect(post!.viewCountUpdated).toBeNull();

  await jobs.updateViewCount({
    postId: post!.id,
  });

  const updated = await models.post.findOne({ id: post.id });
  expect(updated!.viewCount).toEqual(10);
  expect(updated!.viewCountUpdated).not.toBeNull();
});

test("jobs - early return - no value updated", async () => {
  const post = await models.post.create({ title: "My Post" });
  await models.postViews.create({ postId: post.id, views: 3 });
  await models.postViews.create({ postId: post.id, views: 6 });
  await models.postViews.create({ postId: post.id, views: 1 });

  await jobs.updateViewCount({
    postId: "123",
  });

  const updated = await models.post.findOne({ id: post.id });
  expect(updated!.viewCount).toEqual(0);
});

test("jobs - updating all values - values updated", async () => {
  const post1 = await models.post.create({ title: "My First Post" });
  await models.postViews.create({ postId: post1.id, views: 3 });
  await models.postViews.create({ postId: post1.id, views: 6 });
  await models.postViews.create({ postId: post1.id, views: 1 });

  const post2 = await models.post.create({ title: "My Second Post" });
  await models.postViews.create({ postId: post2.id, views: 18 });
  await models.postViews.create({ postId: post2.id, views: 4 });

  const post3 = await models.post.create({ title: "My Third Post" });

  await jobs.updateAllViewCount();

  const updated1 = await models.post.findOne({ id: post1.id });
  expect(updated1!.viewCount).toEqual(10);

  const updated2 = await models.post.findOne({ id: post2.id });
  expect(updated2!.viewCount).toEqual(22);

  const updated3 = await models.post.findOne({ id: post3.id });
  expect(updated3!.viewCount).toEqual(0);
});

test("jobs - updating all values using env var - values updated", async () => {
  const post1 = await models.post.create({ title: "My First Post" });
  await models.postViews.create({ postId: post1.id, views: 3 });
  await models.postViews.create({ postId: post1.id, views: 6 });
  await models.postViews.create({ postId: post1.id, views: 1 });

  const post2 = await models.post.create({ title: "My Second Post" });
  await models.postViews.create({ postId: post2.id, views: 18 });
  await models.postViews.create({ postId: post2.id, views: 4 });

  const post3 = await models.post.create({ title: "My Third Post" });

  await jobs.updateAllViewCount();
  await jobs.updateGoldStarFromEnv();

  const updated1 = await models.post.findOne({ id: post1.id });
  expect(updated1!.status).toEqual(Status.NormalPost);

  const updated2 = await models.post.findOne({ id: post2.id });
  expect(updated2!.status).toEqual(Status.GoldPost);

  const updated3 = await models.post.findOne({ id: post3.id });
  expect(updated3!.status).toEqual(Status.NormalPost);
});

test("jobs - all types as input values", async () => {
  await jobs.allInputTypes({
    text: "text",
    num: 10,
    boolean: true,
    date: new Date(2022, 12, 25),
    timestamp: new Date(2022, 12, 25, 1, 3, 4),
    id: "123",
    enum: Status.GoldPost,
    array: ["one", "two"],
  });
});
