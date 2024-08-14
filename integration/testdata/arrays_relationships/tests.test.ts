import { actions, models, resetDatabase } from "@teamkeel/testing";
import { Genre } from "@teamkeel/sdk";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

test("array relationships - nested creates", async () => {
  const collection = await actions.createCollection({
    name: "Favourites",
    books: [
      {
        title: "Lord of the Rings",
        genres: [Genre.Fantasy, Genre.ScienceFiction],
      },
      {
        title: "Into Thin Air",
        genres: [Genre.Thriller],
      },
    ],
  });

  const books = await models.book.findMany({
    where: { colId: collection.id },
    orderBy: { title: "DESC" },
  });

  expect(books).toHaveLength(2);
  expect(books[0].title).toEqual("Lord of the Rings");
  expect(books[0].genres).toEqual([Genre.Fantasy, Genre.ScienceFiction]);
  expect(books[1].title).toEqual("Into Thin Air");
  expect(books[1].genres).toEqual([Genre.Thriller]);
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
  expect(books[0].genres).toEqual([Genre.All, Genre.New]);
  expect(books[1].title).toEqual("Into Thin Air");
  expect(books[1].genres).toEqual([Genre.All, Genre.New]);
});

test("array relationships - nested query by array equals", async () => {
  const favourites = await actions.createCollection({
    name: "Favourites",
    books: [
      {
        title: "Lord of the Rings",
        genres: [Genre.Fantasy, Genre.ScienceFiction],
      },
      {
        title: "Into Thin Air",
        genres: [Genre.Thriller],
      },
    ],
  });

  const children = await actions.createCollection({
    name: "Children",
    books: [
      {
        title: "Hobbit",
        genres: [Genre.Fantasy, Genre.ScienceFiction],
      },
      {
        title: "The Book of Why",
        genres: [Genre.Science],
      },
    ],
  });

  const collections = await actions.listCollection({
    where: { books: { genres: { equals: [Genre.Science] } } },
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
        genres: [Genre.Fantasy, Genre.ScienceFiction],
      },
      {
        title: "Into Thin Air",
        genres: [Genre.Thriller],
      },
    ],
  });

  const children = await actions.createCollection({
    name: "Children",
    books: [
      {
        title: "Hobbit",
        genres: [Genre.Fantasy, Genre.ScienceFiction],
      },
      {
        title: "The Book of Why",
        genres: [Genre.Science],
      },
    ],
  });

  const maths = await actions.createCollection({
    name: "Education",
    books: [
      {
        title: "Maths 101",
        genres: [Genre.School, Genre.Math],
      },
    ],
  });

  const collections = await actions.listEqualsCollection();
  expect(collections.results).toHaveLength(2);
  expect(collections.results[0].id).toEqual(children.id);
  expect(collections.results[1].id).toEqual(favourites.id);
});

test("array relationships - nested expression array in 1", async () => {
  const favourites = await actions.createCollection({
    name: "Favourites",
    books: [
      {
        title: "Lord of the Rings",
        genres: [Genre.Fantasy, Genre.ScienceFiction],
      },
      {
        title: "Into Thin Air",
        genres: [Genre.Thriller],
      },
    ],
  });

  const children = await actions.createCollection({
    name: "Children",
    books: [
      {
        title: "The Book of Why",
        genres: [Genre.Science],
      },
      {
        title: "Hobbit",
        genres: [Genre.EasyReads, Genre.Fantasy],
      },
    ],
  });

  const maths = await actions.createCollection({
    name: "Education",
    books: [
      {
        title: "Maths 101",
        genres: [Genre.School, Genre.Math],
      },
    ],
  });

  const collections1 = await actions.listInCollection({
    where: { genre: Genre.Fantasy },
  });
  expect(collections1.results).toHaveLength(2);
  expect(collections1.results[0].id).toEqual(children.id);
  expect(collections1.results[1].id).toEqual(favourites.id);

  const collections2 = await actions.listInCollection({
    where: { genre: Genre.Science },
  });
  expect(collections2.results).toHaveLength(1);
  expect(collections2.results[0].id).toEqual(children.id);

  const collections3 = await actions.listInCollection({
    where: { genre: Genre.New },
  });
  expect(collections3.results).toHaveLength(0);
});

// There is a bug with 'not in'. See https://linear.app/keel/issue/KE-2194/arrays-and-not-in
//
// test("array relationships - nested expression array not in", async () => {
//   const favourites = await actions.createCollection({
//     name: "Favourites",
//     books: [
//       {
//         title: "Lord of the Rings",
//         genres: ["fantasy", "science fiction"],
//       },
//       {
//         title: "Into Thin Air",
//         genres: ["thriller"],
//       },
//     ],
//   });

//   const children = await actions.createCollection({
//     name: "Children",
//     books: [
//       {
//         title: "The Book of Why",
//         genres: ["science"],
//       },
//       {
//         title: "Hobbit",
//         genres: ["easy reads", "fantasy"],
//       },
//     ],
//   });

//   const computers = await actions.createCollection({
//     name: "Computers",
//     books: [
//       {
//         title: "C++ For Dummies",
//         genres: ["computer science"],
//       },
//       {
//         title: "PC Master Race",
//         genres: ["technology"],
//       },
//     ],
//   });

//   console.log(await models.collection.findMany());
//   console.log(await models.book.findMany());

//   const collections = await actions.listNotInCollection({ where: {genre: "sdfsd"}});
//   console.log(collections.results);
//   expect(collections.results).toHaveLength(1);
//   expect(collections.results[0].id).toEqual(computers.id);

// });

test("array relationships - nested expression array in 2", async () => {
  const favourites = await actions.createCollection({
    name: "Favourites",
    books: [
      {
        title: "Lord of the Rings",
        genres: [Genre.Fantasy, Genre.ScienceFiction],
      },
      {
        title: "Into Thin Air",
        genres: [Genre.Thriller],
      },
    ],
  });

  const children = await actions.createCollection({
    name: "Children",
    books: [
      {
        title: "Hobbit",
        genres: [Genre.Fantasy, Genre.ScienceFiction],
      },
      {
        title: "The Book of Why",
        genres: [Genre.Science],
      },
    ],
  });

  const identity = await models.identity.create({});
  await models.person.create({
    identityId: identity.id,
    favouriteGenre: Genre.Fantasy,
    favouriteAuthors: ["Tolkien", "Tom"],
  });

  const books = await actions.withIdentity(identity).suggestedBooksByGenre();
  expect(books.results).toHaveLength(2);
  expect(books.results[0].title).toEqual("Hobbit");
  expect(books.results[1].title).toEqual("Lord of the Rings");
});

test("array relationships - nested expression array in 3", async () => {
  const favourites = await actions.createCollection({
    name: "Favourites",
    books: [
      {
        title: "Lord of the Rings",
        genres: [Genre.Fantasy, Genre.ScienceFiction],
        author: "Tolkien",
      },
      {
        title: "Into Thin Air",
        genres: [Genre.Thriller],
        author: "Tom",
      },
    ],
  });

  const children = await actions.createCollection({
    name: "Children",
    books: [
      {
        title: "Hobbit",
        genres: [Genre.Fantasy, Genre.ScienceFiction],
        author: "Tolkien",
      },
      {
        title: "The Book of Why",
        genres: [Genre.Science],
        author: "Dave",
      },
    ],
  });

  const identity = await models.identity.create({});
  await models.person.create({
    identityId: identity.id,
    favouriteGenre: Genre.Fantasy,
    favouriteAuthors: ["Tolkien", "Tom"],
  });

  const books = await actions.withIdentity(identity).suggestedBooksByAuthor();
  expect(books.results).toHaveLength(3);
  expect(books.results[0].title).toEqual("Hobbit");
  expect(books.results[1].title).toEqual("Into Thin Air");
  expect(books.results[2].title).toEqual("Lord of the Rings");
});

test("array relationships - nested expression array in 4", async () => {
  const favourites = await actions.createCollection({
    name: "Favourites",
    books: [
      {
        title: "Lord of the Rings",
        genres: [Genre.Fantasy, Genre.ScienceFiction],
        author: "Tolkien",
      },
      {
        title: "Into Thin Air",
        genres: [Genre.Thriller],
        author: "Tom",
      },
    ],
  });

  const children = await actions.createCollection({
    name: "Children",
    books: [
      {
        title: "Hobbit",
        genres: [Genre.Fantasy, Genre.ScienceFiction],
        author: "Tolkien",
      },
      {
        title: "The Book of Why",
        genres: [Genre.Science],
        author: "Dave",
      },
    ],
  });

  const identity1 = await models.identity.create({});
  await models.person.create({
    identityId: identity1.id,
    favouriteGenre: Genre.Thriller,
    favouriteAuthors: [],
  });

  const collections1 = await actions
    .withIdentity(identity1)
    .suggestedCollections();
  expect(collections1.results).toHaveLength(1);
  expect(collections1.results[0].name).toEqual("Favourites");

  const identity2 = await models.identity.create({});
  await models.person.create({
    identityId: identity2.id,
    favouriteGenre: Genre.Fantasy,
    favouriteAuthors: [],
  });

  const collections2 = await actions
    .withIdentity(identity2)
    .suggestedCollections();
  expect(collections2.results).toHaveLength(2);
  expect(collections2.results[0].name).toEqual("Children");
  expect(collections2.results[1].name).toEqual("Favourites");
});
