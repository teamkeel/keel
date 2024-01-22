import { ProductTagCreateValues, useDatabase } from "@teamkeel/sdk";
import { actions, models, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

test("limit only", async () => {
  await models.post.create({
    title: "a post",
    id: "abc",
  });

  await models.post.create({
    title: "another post",
    id: "def",
  });

  const thirdPost = await models.post.create({
    title: "yet another post",
    id: "ghi",
  });

  const { results: allResults } = await actions.listPosts();

  expect(allResults.length).toEqual(3);

  const { results } = await actions.listPosts({
    where: {
      limit: 2,
    },
  });

  expect(results).not.toContain(thirdPost);
  expect(results.length).toEqual(2);
});

test("offset only", async () => {
  await models.post.create({
    title: "a post",
    id: "abc",
  });

  await models.post.create({
    title: "another post",
    id: "def",
  });

  const third = await models.post.create({
    title: "yet another post",
    id: "ghi",
  });

  const { results } = await actions.listPosts({
    where: {
      offset: 2,
    },
  });

  // if no ordering is specified, then the offset will be applied to
  // ORDER BY posts.id ASC so the third post should only be returned for offset 2
  // as its id is last alphabetically
  expect(results.map((r) => r.id)).toEqual([third.id]);
});

test("order by only", async () => {
  await models.post.create({
    title: "a",
  });
  await models.post.create({
    title: "b",
  });
  await models.post.create({
    title: "c",
  });
  await models.post.create({
    title: "d",
  });

  const { results: ascending } = await actions.listPosts({
    where: {
      orderBy: "title",
      sortOrder: "asc",
    },
  });

  expect(ascending.map((r) => r.title)).toEqual(["a", "b", "c", "d"]);

  const { results: descending } = await actions.listPosts({
    where: {
      orderBy: "title",
      sortOrder: "desc",
    },
  });

  expect(descending.map((r) => r.title)).toEqual(["d", "c", "b", "a"]);
});

test("offset with limit and order by", async () => {
  await createNPosts(10);

  const { results } = await actions.listPosts({
    where: {
      offset: 5,
      limit: 3,
      orderBy: "title",
      sortOrder: "asc",
    },
  });

  const letters = results.map((r) => r.title);

  expect(letters).toEqual(["5", "6", "7"]);
});

test("negative offset", async () => {
  await expect(() =>
    actions.listPosts({
      where: {
        offset: -1,
        limit: 3,
        orderBy: "title",
        sortOrder: "asc",
      },
    })
  ).rejects.toThrow("OFFSET must not be negative");
});

test("unknown sort column", async () => {
  await expect(() =>
    actions.listPosts({
      where: {
        offset: 5,
        limit: 3,
        orderBy: "foo",
        sortOrder: "asc",
      },
    })
  ).rejects.toThrow("column post.foo does not exist");
});

const createNPosts = async (n: number) =>
  Promise.all(
    Array.from(Array(n).keys()).map(async (i) => {
      return models.post.create({
        title: i.toString(),
      });
    })
  );

test("findOne - compound unique with relationships", async () => {
  const tom = await models.profile.create({
    name: "Tom",
  });
  const benoit = await models.profile.create({
    name: "Benoit",
  });

  await models.follow.create({
    fromId: tom.id,
    toId: benoit.id,
  });

  const dbFollow = await models.follow.findOne({
    fromId: tom.id,
    toId: benoit.id,
  });

  expect(dbFollow).not.toBeNull();
});

test("findOne - has-many", async () => {
  const dickens = await models.author.create({
    name: "Charles Dickens",
  });
  const book = await models.book.create({
    authorId: dickens.id,
    title: "Oliver Twist",
  });

  const dbAuthor = await models.author.findOne({
    books: {
      id: book.id,
    },
  });

  expect(dbAuthor).not.toBeNull();
});

test("findOne - one-to-one", async () => {
  const u = await models.user.create({});
  const s = await models.settings.create({
    userId: u.id,
  });

  const dbSettings = await models.settings.findOne({
    userId: u.id,
  });
  expect(dbSettings).not.toBeNull();
  expect(dbSettings!.id).toEqual(s.id);
});

test("create - nested many-to-many with existing records", async () => {
  const [tag1, tag2] = await Promise.all([
    models.tag.create({
      tag: "big",
    }),
    models.tag.create({
      tag: "yellow",
    }),
  ]);

  const bigBird = await models.product.create({
    title: "Big Bird",
    tags: [
      {
        tagId: tag1.id,
      },
      {
        tagId: tag2.id,
      },
    ],
  });

  const dbBigBird = await models.product.findOne({
    id: bigBird.id,
  });
  expect(dbBigBird).not.toBeNull();
  expect(dbBigBird?.title).toBe("Big Bird");

  const dbTags = await models.productTag.findMany({
    where: {
      productId: dbBigBird?.id,
    },
  });
  expect(dbTags.length).toBe(2);

  const tagIds = dbTags.map((x) => x.tagId);
  expect(tagIds).toContain(tag1.id);
  expect(tagIds).toContain(tag2.id);
});

test("create - nested many-to-many with new records", async () => {
  const buzz = await models.product.create({
    title: "Buzz Lightyear",
    tags: [
      {
        tag: {
          tag: "infinity",
        },
      },
      {
        tag: {
          tag: "beyond",
        },
      },
    ],
  });

  const dbBuzz = await models.product.findOne({
    id: buzz.id,
  });
  expect(dbBuzz).not.toBeNull();
  expect(dbBuzz?.title).toBe("Buzz Lightyear");

  const dbTags = await useDatabase()
    .selectFrom("tag")
    .selectAll()
    .whereExists((qb) => {
      return qb
        .selectFrom("product_tag")
        .select("id")
        .where("productId", "=", dbBuzz!.id);
    })
    .execute();

  expect(dbTags.length).toBe(2);

  const tagValues = dbTags.map((x) => x.tag);
  expect(tagValues).toContain("infinity");
  expect(tagValues).toContain("beyond");
});

test("create - nested has-many two-levels", async () => {
  const course = await models.course.create({
    title: "Computer Science",
    lessons: [
      {
        title: "How to be a senior engineer",
        readings: [
          {
            book: "It depends",
            fromPage: 1,
            toPage: 10,
          },
        ],
      },
      {
        title: "How to get fired",
        readings: [
          {
            book: "Rewriting your app in C over the weekend for fun and profit",
            fromPage: 100,
            toPage: 101,
          },
          {
            book: "Who needs <a> when you can use <div> with onclick?",
            fromPage: 10,
            toPage: 14,
          },
        ],
      },
      {
        title: "Data structures",
        readings: [
          {
            book: "Everything is just a hashmap in the end",
            fromPage: 100,
            toPage: 101,
          },
        ],
      },
    ],
  });

  const dbLessons = await models.lesson.findMany({
    where: {
      courseId: course.id,
    },
  });
  expect(dbLessons.length).toBe(3);

  const dbReadings = await models.reading.findMany({
    where: {
      lessonId: {
        oneOf: dbLessons.map((x) => x.id),
      },
    },
  });
  expect(dbReadings.length).toBe(4);
});

test("findMany - notEquals", async () => {
  const dickens = await models.author.create({
    name: "Charles Dickens",
  });
  const tolkien = await models.author.create({
    name: "J.R.R. Tolkien",
  });

  const authors = await models.author.findMany({
    where: {
      id: {
        notEquals: dickens.id,
      },
    },
  });

  expect(authors.length).toBe(1);
  expect(authors[0].id).equals(tolkien.id);
});
