import { models, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

test("create", async () => {
  const post = await models.post.create({ title: "apple" });

  expect(post.title).toEqual("apple");
});

test("update", async () => {
  const post = await models.post.create({ title: "star wars" });

  const updatedPost = await models.post.update(
    { id: post.id },
    {
      title: "star wars sucks!",
    }
  );

  expect(updatedPost.title).toEqual("star wars sucks!");
});

test("findMany", async () => {
  await models.post.create({ title: "apple" });
  await models.post.create({ title: "apple pie" });
  await models.post.create({ title: "pear" });

  const results = await models.post.findMany({
    where: {
      title: {
        startsWith: "apple",
      },
    },
  });

  expect(results.length).toEqual(2);
});

test("findOne", async () => {
  const post = await models.post.create({ title: "ghi" });
  await models.post.create({ title: "hij" });

  const { id } = post;

  const p = await models.post.findOne({ id });
  expect(p!.id).toEqual(post.id);
});
