import { actions, models, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

test("create action", async () => {
  const date = new Date();

  const createdPost = await actions.createPost({
    text: "foo",
    date: date,
    dateTime: date,
  });

  expect(createdPost.text).toEqual("foo");
  expect(createdPost.date).toEqual(date.toISOString().slice(0, 10));
  expect(createdPost.dateTime.toISOString()).toEqual(date.toISOString());
});

test("get action", async () => {
  const date = new Date();
  const post = await actions.createPost({
    text: "foo",
    date: date,
    dateTime: date,
  });

  const fetchedPost = await actions.getPost({ id: post.id });
  expect(fetchedPost).not.toBeNull();
  expect(fetchedPost!.text).toEqual("foo");
  expect(fetchedPost!.date).toEqual(date.toISOString().slice(0, 10));
  expect(fetchedPost!.dateTime.toISOString()).toEqual(date.toISOString());
});
