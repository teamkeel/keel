import { actions, models, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

test("array relationships - nested creates", async () => {
  const collection = await actions.createCollection({
    name: "Favourites",
    books: [
      {
        title: "Lord of the Rings",
        genres: ["fantasy", "science fiction"],
      },
      {
        title: "Into Thin Air",
        genres: ["thriller"],
      },
    ],
  });

  const books = await models.book.findMany({
    where: { colId: collection.id },
    orderBy: { title: "DESC" },
  });

  expect(books).toHaveLength(2);
  expect(books[0].title).toEqual("Lord of the Rings");
  expect(books[0].genres).toEqual(["fantasy", "science fiction"]);
  expect(books[1].title).toEqual("Into Thin Air");
  expect(books[1].genres).toEqual(["thriller"]);
});

test("array relationships - nested creates with @set", async () => {
  const collection = await actions.createCollectionSetGenres({
    name: "Favourites",
    books: [
      {
        title: "Lord of the Rings",
      },
      {
        title: "Into Thin Air",
      },
    ],
  });

  const books = await models.book.findMany({
    where: { colId: collection.id },
    orderBy: { title: "DESC" },
  });

  expect(books).toHaveLength(2);
  expect(books[0].title).toEqual("Lord of the Rings");
  expect(books[0].genres).toEqual(["all", "new"]);
  expect(books[1].title).toEqual("Into Thin Air");
  expect(books[1].genres).toEqual(["all", "new"]);
});

test("array relationships - nested query by array equals", async () => {
  const favourites = await actions.createCollection({
    name: "Favourites",
    books: [
      {
        title: "Lord of the Rings",
        genres: ["fantasy", "science fiction"],
      },
      {
        title: "Into Thin Air",
        genres: ["thriller"],
      },
    ],
  });

  const children = await actions.createCollection({
    name: "Children",
    books: [
      {
        title: "Hobbit",
        genres: ["fantasy", "science fiction"],
      },
      {
        title: "The Book of Why",
        genres: ["science"],
      },
    ],
  });

  const collections = await actions.listCollection({
    where: { books: { genres: { equals: ["science"] } } },
  });
  expect(collections.results).toHaveLength(1);
  expect(collections.results[0].id).toEqual(children.id);
});

test("array relationships - nested expression array equals", async () => {
  const favourites = await actions.createCollection({
    name: "Favourites",
    books: [
      {
        title: "Lord of the Rings",
        genres: ["fantasy", "science fiction"],
      },
      {
        title: "Into Thin Air",
        genres: ["thriller"],
      },
    ],
  });

  const children = await actions.createCollection({
    name: "Children",
    books: [
      {
        title: "Hobbit",
        genres: ["fantasy", "science fiction"],
      },
      {
        title: "The Book of Why",
        genres: ["science"],
      },
    ],
  });

  const maths = await actions.createCollection({
    name: "Education",
    books: [
      {
        title: "Maths 101",
        genres: ["school", "math"],
      },
    ],
  });

  const collections = await actions.listEqualsCollection();
  expect(collections.results).toHaveLength(2);
  expect(collections.results[0].id).toEqual(children.id);
  expect(collections.results[1].id).toEqual(favourites.id);
});

test("array relationships - nested expression array in", async () => {
  const favourites = await actions.createCollection({
    name: "Favourites",
    books: [
      {
        title: "Lord of the Rings",
        genres: ["fantasy", "science fiction"],
      },
      {
        title: "Into Thin Air",
        genres: ["thriller"],
      },
    ],
  });

  const children = await actions.createCollection({
    name: "Children",
    books: [
      {
        title: "Hobbit",
        genres: ["fantasy", "science fiction"],
      },
      {
        title: "The Book of Why",
        genres: ["science"],
      },
    ],
  });

  const maths = await actions.createCollection({
    name: "Education",
    books: [
      {
        title: "Maths 101",
        genres: ["school", "math"],
      },
    ],
  });

  const collections = await actions.listEqualsCollection();
  expect(collections.results).toHaveLength(2);
  expect(collections.results[0].id).toEqual(children.id);
  expect(collections.results[1].id).toEqual(favourites.id);
});

test("array relationships - nested expression array not in", async () => {
  const favourites = await actions.createCollection({
    name: "Favourites",
    books: [
      {
        title: "Lord of the Rings",
        genres: ["fantasy", "science fiction"],
      },
      {
        title: "Into Thin Air",
        genres: ["thriller"],
      },
    ],
  });

  const children = await actions.createCollection({
    name: "Children",
    books: [
      {
        title: "Hobbit",
        genres: ["fantasy", "science fiction"],
      },
      {
        title: "The Book of Why",
        genres: ["science"],
      },
    ],
  });

  const collections = await actions.listEqualsCollection();
  expect(collections.results).toHaveLength(2);
  expect(collections.results[0].id).toEqual(children.id);
  expect(collections.results[1].id).toEqual(favourites.id);
});
