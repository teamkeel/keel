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
