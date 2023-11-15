import { actions, resetDatabase, models } from "@teamkeel/testing";
import { beforeEach, expect, test } from "vitest";

beforeEach(resetDatabase);

test("create beforeWrite - mutate values", async () => {
  const book = await actions.createBookBeforeWrite({
    title: "Great Gatsby",
  });

  expect(book.title).toEqual("GREAT GATSBY");

  const dbBook = await models.book.findOne({
    id: book.id,
  });
  expect(dbBook).not.toBeNull();
  expect(dbBook!.title).toEqual("GREAT GATSBY");
});

test("create beforeWrite - mutate values sync", async () => {
  const book = await actions.createBookBeforeWriteSync({
    title: "Great Gatsby",
  });

  expect(book.title).toEqual("GREAT GATSBY");

  const dbBook = await models.book.findOne({
    id: book.id,
  });
  expect(dbBook).not.toBeNull();
  expect(dbBook!.title).toEqual("GREAT GATSBY");
});

test("create afterWrite - create additional records", async () => {
  const book = await actions.createBookAfterWrite({
    title: "Robinson Crusoe",
    review: "This is a great book",
  });

  const reviews = await models.review.findMany({
    where: {
      bookId: book.id,
    },
  });
  expect(reviews.length).toEqual(1);
  expect(reviews[0].review).toEqual("This is a great book");
});

test("create afterWrite - error and rollback", async () => {
  expect(
    actions.createBookAfterWriteErrorRollback({
      title: "Lady Chatterley's Lover",
    })
  ).rejects.toEqual({
    code: "ERR_INTERNAL",
    message: "this book is banned",
  });

  // Check the book was not created
  const books = await models.book.findMany({
    where: {
      title: {
        equals: "Lady Chatterley's Lover",
      },
    },
  });
  expect(books.length).toEqual(0);
});

test("get beforeQuery - return null", async () => {
  const book = await actions.getBookBeforeQueryFirstOrNull({
    title: "This book doesnt exist",
  });
  expect(book).toBeNull();
});

test("get beforeQuery - returns Promise<Book>", async () => {
  const dbBook = await models.book.create({
    title: "This book does exist",
  });
  const book = await actions.getBookBeforeQueryFirstOrNull({
    title: "This book does exist",
  });
  expect(book).not.toBeNull();
  expect(book!.id).toEqual(dbBook.id);
});

test("get beforeQuery - returns QueryBuilder", async () => {
  const dbBook = await models.book.create({
    title: "A great book",
    published: false,
  });
  let book = await actions.getBookBeforeQueryQueryBuilder({
    id: dbBook.id,
  });
  expect(book).toBeNull();

  book = await actions.getBookBeforeQueryQueryBuilder({
    id: dbBook.id,
    allowUnpublished: true,
  });
  expect(book).not.toBeNull();
  expect(book!.id).toEqual(dbBook.id);
});

test("get afterQuery - mutate returned data", async () => {
  const dbBook = await models.book.create({
    title: "Why crypto is the future",
  });
  let book = await actions.getBookAfterQuery({
    id: dbBook.id,
  });

  expect(book).not.toBeNull();
  expect(book!.id).toEqual(dbBook.id);
  // Returned data should have been mutated by the hook
  expect(book!.title).toEqual("Why c****o is the future");

  // Database record should not have changed
  const dbBook2 = await models.book.findOne({
    id: dbBook.id,
  });
  expect(book).not.toBeNull();
  expect(dbBook2!.title).toEqual("Why crypto is the future");
});

test("get afterQuery - permission denied", async () => {
  const dbBook = await models.book.create({
    title: "Star Wars X - Ja Ja's Back",
    published: false,
  });

  expect(
    actions.getBookAfterQueryPermissions({
      id: dbBook.id,
      onlyPublished: true,
    })
  ).rejects.toEqual({
    code: "ERR_PERMISSION_DENIED",
    message: "not authorized to access this action",
  });
});

test("list beforeQuery - updated QueryBuilder", async () => {
  await models.book.create({
    title: "Practical Magic",
  });
  const dbBook = await models.book.create({
    title: "The Colour of Magic",
  });
  await models.book.create({
    title: "The Magic Mountain",
  });

  //
  const books = await actions.listBooksBeforeQuery({
    where: {
      title: {
        startsWith: "The",
      },
    },
  });

  expect(books.results.length).toEqual(1);
  expect(books.results[0].id).toEqual(dbBook.id);
});

test("list beforeQuery - with first", async () => {
  await models.book.create({
    title: "Practical Magic",
  });
  await models.book.create({
    title: "The Colour of Magic",
  });
  await models.book.create({
    title: "The Rules of Magic",
  });
  await models.book.create({
    title: "The Magic Mountain",
  });

  // There are three matching books but we ask for only the first two
  const books = await actions.listBooksBeforeQuery({
    first: 2,
  });

  expect(books.results.length).toEqual(2);
});

test("list beforeQuery - return values", async () => {
  const books = await actions.listBooksBeforeQueryReturnValues();

  expect(books.results.length).toEqual(1);
  expect(books.results[0]).toEqual({
    id: "1234",
    createdAt: new Date("2001-01-01"),
    updatedAt: new Date("2001-01-01"),
    title: "Dreamcatcher",
    published: true,
  });
});

test("list afterQuery - mutate values", async () => {
  const lotr = await models.book.create({
    title: "The Lord of the Rings",
  });
  const hobbit = await models.book.create({
    title: "The Hobbit",
  });

  const books = await actions.listBooksAfterQuery();

  expect(books.results.length).toEqual(2);
  const titles = books.results.map((x) => x.title);
  titles.sort();

  // Check returned value have been mutated
  expect(titles).toEqual(["THE HOBBIT", "THE LORD OF THE RINGS"]);

  // Check records in the database should not have changed
  expect((await models.book.findOne({ id: lotr.id }))?.title).toEqual(
    "The Lord of the Rings"
  );
  expect((await models.book.findOne({ id: hobbit.id }))?.title).toEqual(
    "The Hobbit"
  );
});

test("list afterQuery - permission denied", async () => {
  await models.book.create({
    title: "Lady Chatterley's Lover",
    published: false,
  });
  await models.book.create({
    title: "Dark Lover",
    published: true,
  });

  await expect(
    actions.listBooksAfterQueryPermissions({
      where: {
        onlyPublished: true,
      },
    })
  ).rejects.toEqual({
    code: "ERR_PERMISSION_DENIED",
    message: "not authorized to access this action",
  });
});

test("update beforeQuery - returning QueryBuilder no record", async () => {
  const book = await models.book.create({
    title: "Lady Chatterley's Lover",
    published: false,
  });

  await expect(
    actions.updateBookBeforeQuery({
      where: {
        id: book.id,
        returnRecord: false,
      },
      values: {
        title: "my new title",
      },
    })
  ).rejects.toEqual({
    code: "ERR_RECORD_NOT_FOUND",
    message: "record not found",
  });
});

test("update beforeQuery - no record", async () => {
  const book = await models.book.create({
    title: "Lady Chatterley's Lover",
    published: false,
  });

  await expect(
    actions.updateBookBeforeQuery({
      where: {
        id: book.id,
        returnRecord: true,
      },
      values: {
        title: "my new title",
      },
    })
  ).rejects.toEqual({
    code: "ERR_RECORD_NOT_FOUND",
    message: "record not found",
  });
});

test("update beforeWrite - use existing record values", async () => {
  const dbBook = await models.book.create({
    title: "Harry Potter",
    published: false,
  });

  let book = await actions.updateBookBeforeWrite({
    where: {
      id: dbBook.id,
    },
    values: {
      title: "my new title",
    },
  });

  expect(book.title).toEqual("my new title");
  expect(book.published).toEqual(true);
});

test("update beforeWrite - permission denied", async () => {
  const dbBook = await models.book.create({
    title: "Harry Potter",
    published: false,
  });

  await expect(
    actions.updateBookBeforeWrite({
      where: {
        id: dbBook.id,
      },
      values: {
        title: "How to Build a Bomb in 10 Easy Steps",
      },
    })
  ).rejects.toEqual({
    code: "ERR_PERMISSION_DENIED",
    message: "not authorized to access this action",
  });
});

test("update afterWrite - create/update additional records", async () => {
  const dbBook = await models.book.create({
    title: "Harry Potter",
  });

  let book = await actions.updateBookAfterWrite({
    where: {
      id: dbBook.id,
    },
    values: {
      title: "my new title",
    },
  });

  expect(book.title).toEqual("MY NEW TITLE");

  let updates = await models.bookUpdates.findOne({
    bookId: book.id,
  });
  expect(updates!.updateCount).toEqual(1);

  book = await actions.updateBookAfterWrite({
    where: {
      id: dbBook.id,
    },
    values: {
      title: "my different title",
    },
  });

  expect(book.title).toEqual("MY DIFFERENT TITLE");

  updates = await models.bookUpdates.findOne({
    bookId: book.id,
  });
  expect(updates!.updateCount).toEqual(2);
});

test("delete beforeQuery - mutate query", async () => {
  const dbBook = await models.book.create({
    title: "Harry Potter",
    published: true,
  });

  let bookId = await actions.deleteBookBeforeQuery({
    id: dbBook.id,
    allowPublished: true,
  });

  expect(bookId).toEqual(dbBook.id);

  const b = await models.book.findOne({ id: dbBook.id });
  expect(b).toBeNull();
});

test("delete beforeQuery - return record", async () => {
  const dbBook = await models.book.create({
    title: "Harry Potter",
    published: true,
  });

  let bookId = await actions.deleteBookBeforeQueryReturnRecord({
    id: dbBook.id,
  });

  expect(bookId).toEqual(dbBook.id);

  const b = await models.book.findOne({ id: dbBook.id });
  expect(b).toBeNull();
});

test("delete beforeQuery - mutate query not found", async () => {
  const dbBook = await models.book.create({
    title: "Harry Potter",
    published: true,
  });

  await expect(
    actions.deleteBookBeforeQuery({
      id: dbBook.id,
      allowPublished: false,
    })
  ).rejects.toEqual({
    code: "ERR_RECORD_NOT_FOUND",
    message: "record not found",
  });
});

test("delete beforeWrite - permission denied", async () => {
  let dbBook = await models.book.create({
    title: "Harry Potter",
    published: true,
  });

  await expect(
    actions.deleteBookBeforeWrite({
      id: dbBook.id,
      allowPublished: false,
    })
  ).rejects.toEqual({
    code: "ERR_PERMISSION_DENIED",
    message: "not authorized to access this action",
  });

  const dbBook2 = await models.book.findOne({
    id: dbBook.id,
  });
  expect(dbBook2).not.toBeNull();
});

test("delete beforeWrite - create record", async () => {
  let dbBook = await models.book.create({
    title: "Harry Potter",
    published: true,
  });

  await actions.deleteBookBeforeWrite({
    id: dbBook.id,
    allowPublished: true,
  });

  const dbBook2 = await models.book.findOne({
    id: dbBook.id,
  });
  expect(dbBook2).toBeNull();

  const deletedBooks = await models.deletedBook.findMany({
    where: {
      bookId: {
        equals: dbBook.id,
      },
    },
  });
  expect(deletedBooks.length).toEqual(1);
  expect(deletedBooks[0].title).toEqual("Harry Potter");
  expect(deletedBooks[0].bookId).toEqual(dbBook.id);
});

test("delete afterWrite - create record", async () => {
  let dbBook = await models.book.create({
    title: "Anna Karenina",
  });

  await actions.deleteBookAfterWrite({
    id: dbBook.id,
    reason: "too long",
  });

  const dbBook2 = await models.book.findOne({
    id: dbBook.id,
  });
  expect(dbBook2).toBeNull();

  const deletedBooks = await models.deletedBook.findMany({
    where: {
      bookId: {
        equals: dbBook.id,
      },
    },
  });
  expect(deletedBooks.length).toEqual(1);
  expect(deletedBooks[0].title).toEqual("Anna Karenina (too long)");
  expect(deletedBooks[0].bookId).toEqual(dbBook.id);
});
