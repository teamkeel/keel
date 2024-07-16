import { models, jobs, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";
import { Status, InlineFile } from "@teamkeel/sdk";

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

  const dataUrl = `data:image/png;name=my-avatar.png;base64,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`;
  
  await jobs.allInputTypes({
    text: "text",
    num: 10,
    boolean: true,
    date: new Date(2022, 12, 25),
    timestamp: new Date(2022, 12, 25, 1, 3, 4),
    id: "123",
    enum: Status.GoldPost,
    image: InlineFile.fromDataURL(dataUrl),
  });
});
