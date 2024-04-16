import { test, expect, beforeEach, beforeAll } from "vitest";
import { actions, resetDatabase, models } from "@teamkeel/testing";

beforeAll(async () => {
  await models.item.create({
    value: 1,
    letter: "a",
    name: "a",
  });
  await models.item.create({
    value: 2,
    letter: "b",
    name: "bee",
  });
  await models.item.create({
    value: 3,
    letter: "c",
    name: "cee",
  });
  await models.item.create({
    value: 4,
    letter: "d",
    name: "dee",
  });
  await models.item.create({
    value: 5,
    letter: "e",
    name: "e",
  });
  await models.item.create({
    value: 6,
    letter: "f",
    name: "ef",
  });
  await models.item.create({
    value: 7,
    letter: "g",
    name: "gee",
  });
  await models.item.create({
    value: 8,
    letter: "h",
    name: "(h)aitch",
  });
  await models.item.create({
    value: 9,
    letter: "i",
    name: "i",
  });
  await models.item.create({
    value: 10,
    letter: "j",
    name: "jay",
  });
  await models.item.create({
    value: 11,
    letter: "k",
    name: "kay",
  });
  await models.item.create({
    value: 12,
    letter: "l",
    name: "el",
  });
  await models.item.create({
    value: 13,
    letter: "m",
    name: "em",
  });
  await models.item.create({
    value: 14,
    letter: "n",
    name: "en",
  });
  await models.item.create({
    value: 15,
    letter: "o",
    name: "o",
  });
  await models.item.create({
    value: 16,
    letter: "p",
    name: "pee",
  });
  await models.item.create({
    value: 17,
    letter: "q",
    name: "cue",
  });
  await models.item.create({
    value: 18,
    letter: "r",
    name: "ar",
  });
  await models.item.create({
    value: 19,
    letter: "s",
    name: "ess",
  });
  await models.item.create({
    value: 20,
    letter: "t",
    name: "tee",
  });
  await models.item.create({
    value: 21,
    letter: "u",
    name: "u",
  });
  await models.item.create({
    value: 22,
    letter: "v",
    name: "vee",
  });
  await models.item.create({
    value: 23,
    letter: "w",
    name: "double-u",
  });
  await models.item.create({
    value: 24,
    letter: "x",
    name: "ex",
  });
  await models.item.create({
    value: 25,
    letter: "y",
    name: "wy",
  });
  await models.item.create({
    value: 26,
    letter: "z",
    name: "zee/zed",
  });
});

test("pagination - first page", async () => {
  const alphabet = await actions.listItems({
    orderBy: [{ letter: "asc" }],
    first: 3,
  });

  expect(alphabet.pageInfo.count).toEqual(3);
  expect(alphabet.pageInfo.totalCount).toEqual(26);
  expect(alphabet.pageInfo.startCursor).toEqual(alphabet.results[0].id);
  expect(alphabet.pageInfo.endCursor).toEqual(alphabet.results[2].id);
  expect(alphabet.pageInfo.hasNextPage).toEqual(true);
  expect(alphabet.results[0].letter).toEqual("a");
  expect(alphabet.results[1].letter).toEqual("b");
  expect(alphabet.results[2].letter).toEqual("c");
});

test("pagination - forward", async () => {
  // get first page of 10
  const alphabet = await actions.listItems({
    orderBy: [{ value: "asc" }],
    first: 10,
  });

  expect(alphabet.pageInfo.count).toEqual(10);
  expect(alphabet.pageInfo.totalCount).toEqual(26);
  expect(alphabet.pageInfo.startCursor).toEqual(alphabet.results[0].id);
  expect(alphabet.pageInfo.endCursor).toEqual(alphabet.results[9].id);
  expect(alphabet.pageInfo.hasNextPage).toEqual(true);
  expect(alphabet.results[0].letter).toEqual("a");
  expect(alphabet.results[9].letter).toEqual("j");

  // get second page of 10
  const alphabet2 = await actions.listItems({
    orderBy: [{ value: "asc" }],
    first: 10,
    after: alphabet.pageInfo.endCursor,
  });

  expect(alphabet2.pageInfo.count).toEqual(10);
  expect(alphabet2.pageInfo.totalCount).toEqual(26);
  expect(alphabet2.pageInfo.startCursor).toEqual(alphabet2.results[0].id);
  expect(alphabet2.pageInfo.endCursor).toEqual(alphabet2.results[9].id);
  expect(alphabet2.pageInfo.hasNextPage).toEqual(true);
  expect(alphabet2.results[0].letter).toEqual("k");
  expect(alphabet2.results[9].letter).toEqual("t");

  // get last page
  const alphabet3 = await actions.listItems({
    orderBy: [{ value: "asc" }],
    first: 10,
    after: alphabet2.pageInfo.endCursor,
  });

  expect(alphabet3.pageInfo.count).toEqual(6);
  expect(alphabet3.pageInfo.totalCount).toEqual(26);
  expect(alphabet3.pageInfo.startCursor).toEqual(alphabet3.results[0].id);
  expect(alphabet3.pageInfo.endCursor).toEqual(alphabet3.results[5].id);
  expect(alphabet3.pageInfo.hasNextPage).toEqual(false);
  expect(alphabet3.results[0].letter).toEqual("u");
  expect(alphabet3.results[5].letter).toEqual("z");
});

test("pagination - backwards", async () => {
  const allItems = await actions.listItems({
    orderBy: [{ value: "asc" }],
    first: 100,
  });

  expect(allItems.results[25].letter).toEqual("z");

  // get last page of 10
  const alphabet = await actions.listItems({
    orderBy: [{ value: "asc" }],
    last: 10,
    before: allItems.results[25].id,
  });

  expect(alphabet.pageInfo.count).toEqual(10);
  expect(alphabet.pageInfo.totalCount).toEqual(26);
  expect(alphabet.pageInfo.startCursor).toEqual(alphabet.results[0].id);
  expect(alphabet.pageInfo.endCursor).toEqual(alphabet.results[9].id);
  expect(alphabet.pageInfo.hasNextPage).toEqual(false);
  expect(alphabet.results[0].value).toEqual(16);
  expect(alphabet.results[9].value).toEqual(25);

  // get second page of 10
  const alphabet2 = await actions.listItems({
    orderBy: [{ value: "asc" }],

    last: 10,
    before: alphabet.pageInfo.startCursor,
  });

  expect(alphabet2.pageInfo.count).toEqual(10);
  expect(alphabet2.pageInfo.totalCount).toEqual(26);
  expect(alphabet2.pageInfo.startCursor).toEqual(alphabet2.results[0].id);
  expect(alphabet2.pageInfo.endCursor).toEqual(alphabet2.results[9].id);
  expect(alphabet2.pageInfo.hasNextPage).toEqual(false);
  expect(alphabet2.results[0].value).toEqual(6);
  expect(alphabet2.results[9].value).toEqual(15);

  // get last page
  const alphabet3 = await actions.listItems({
    orderBy: [{ value: "asc" }],
    last: 10,
    before: alphabet2.pageInfo.startCursor,
  });

  expect(alphabet3.pageInfo.count).toEqual(5);
  expect(alphabet3.pageInfo.totalCount).toEqual(26);
  expect(alphabet3.pageInfo.startCursor).toEqual(alphabet3.results[0].id);
  expect(alphabet3.pageInfo.endCursor).toEqual(alphabet3.results[4].id);
  expect(alphabet3.pageInfo.hasNextPage).toEqual(false);
  expect(alphabet3.results[0].value).toEqual(1);
  expect(alphabet3.results[4].value).toEqual(5);
});
