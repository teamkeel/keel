import { test, expect, beforeEach } from "vitest";
import { actions, resetDatabase } from "@teamkeel/testing";

beforeEach(resetDatabase);

test("List Where filters - using equal operator (string) - filters correctly", async () => {
  await actions.createPost({ title: "Fred" });
  await actions.createPost({ title: "NotFred" });

  const { results } = await actions.listPostsEqualString({
    where: {
      whereArg: "Fred",
    },
  });

  expect(results.length).toEqual(1);
  expect(results[0].title).toEqual("Fred");
});

test("List Where filters inverse - using equal operator (string) - filters correctly", async () => {
  await actions.createPost({ title: "Fred" });
  await actions.createPost({ title: "NotFred" });

  const { results } = await actions.listPostsEqualStringInverse({
    where: {
      whereArg: "Fred",
    },
  });

  expect(results.length).toEqual(1);
  expect(results[0].title).toEqual("Fred");
});

test("List Where filters - using not equal operator (string) - filters correctly", async () => {
  await actions.createPost({ title: "Fred" });
  await actions.createPost({ title: "NotFred" });

  const { results } = await actions.listPostsNotEqualString({
    where: {
      whereArg: "Fred",
    },
  });

  expect(results.length).toEqual(1);
  expect(results[0].title).toEqual("NotFred");
});

test("List Where filters - using equal operator on date - filters correctly", async () => {
  await actions.createPost({ aDate: new Date(2020, 1, 21) });
  await actions.createPost({ aDate: new Date(2020, 1, 22) });
  await actions.createPost({ aDate: new Date(2020, 1, 23) });

  const { results } = await actions.listPostsEqualDate({
    where: {
      whereArg: new Date(2020, 1, 21),
    },
  });

  expect(results.length).toEqual(1);
});

test("List Where filters - using not equal operator on date - filters correctly", async () => {
  await actions.createPost({ aDate: new Date(2020, 1, 21) });
  await actions.createPost({ aDate: new Date(2020, 1, 22) });
  await actions.createPost({ aDate: new Date(2020, 1, 23) });

  const { results } = await actions.listPostsNotEqualDate({
    where: {
      whereArg: new Date(2020, 1, 21),
    },
  });

  expect(results.length).toEqual(2);
});

test("List Where filters - using after operator on timestamp - filters correctly", async () => {
  await actions.createPost({ aTimestamp: new Date(2020, 1, 21, 1, 0, 0) });
  await actions.createPost({ aTimestamp: new Date(2020, 1, 22, 2, 30, 0) });
  await actions.createPost({ aTimestamp: new Date(2020, 1, 23, 4, 0, 0) });

  const { results } = await actions.listPostsAfterTimestamp({
    where: {
      whereArg: new Date(2020, 1, 21, 1, 30, 0),
    },
  });

  expect(results.length).toEqual(2);
});

test("List Where filters - using before operator on timestamp - filters correctly", async () => {
  await actions.createPost({ aTimestamp: new Date(2020, 1, 21, 1, 0, 0) });
  await actions.createPost({ aTimestamp: new Date(2020, 1, 22, 2, 30, 0) });
  await actions.createPost({ aTimestamp: new Date(2020, 1, 23, 4, 0, 0) });

  const { results } = await actions.beforePostsBeforeTimestamp({
    where: {
      whereArg: new Date(2020, 1, 21, 1, 30, 0),
    },
  });

  expect(results.length).toEqual(1);
});

test("List Where filters - using less than operator on decimal - filters correctly", async () => {
  await actions.createPost({ aDecimal: 1.5 });
  await actions.createPost({ aDecimal: 2.6 });
  await actions.createPost({ aDecimal: 3.7 });

  const { results } = await actions.listPostsLessThanDecimal({
    where: {
      whereArg: 3.5,
    },
  });

  expect(results.length).toEqual(2);
});

test("List Where filters - using greater than operator on decimal - filters correctly", async () => {
  await actions.createPost({ aDecimal: 1.5 });
  await actions.createPost({ aDecimal: 2.6 });
  await actions.createPost({ aDecimal: 3.7 });

  const { results } = await actions.listPostsGreaterThanDecimal({
    where: {
      whereArg: 1.5,
    },
  });

  expect(results.length).toEqual(2);
});

test("List Where filters - using equal operator on decimal - filters correctly", async () => {
  await actions.createPost({ aDecimal: 1.5 });

  const { results } = await actions.listPostsEqualDecimal({
    where: {
      whereArg: 1.5,
    },
  });

  expect(results.length).toEqual(1);
});
