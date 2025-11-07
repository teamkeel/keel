import { test, expect, beforeEach, describe } from "vitest";
import { ModelAPI } from "./ModelAPI";
import { sql } from "kysely";
import { useDatabase } from "./database";
import KSUID from "ksuid";

let personAPI;
let postAPI;
let authorAPI;

beforeEach(async () => {
  const db = useDatabase();

  await sql`
  DROP TABLE IF EXISTS post;
  DROP TABLE IF EXISTS person;
  DROP TABLE IF EXISTS author;

  CREATE TABLE person(
      id               text PRIMARY KEY,
      name             text UNIQUE,
      married          boolean,
      favourite_number integer,
      avatar           jsonb,
      date             timestamp
  );
  CREATE TABLE post(
    id               text PRIMARY KEY,
    title            text,
    tags             text[],
    rating           numeric,
    author_id        text references person(id)
  );
  CREATE TABLE author(
    id               text PRIMARY KEY,
    name             text NOT NULL
  );`.execute(db);

  const tableConfigMap = {
    person: {
      posts: {
        relationshipType: "hasMany",
        foreignKey: "author_id",
        referencesTable: "post",
      },
    },
    post: {
      author: {
        relationshipType: "belongsTo",
        foreignKey: "author_id",
        referencesTable: "person",
      },
    },
  };

  personAPI = new ModelAPI("person", undefined, tableConfigMap);

  postAPI = new ModelAPI("post", undefined, tableConfigMap);

  authorAPI = new ModelAPI("author", undefined, tableConfigMap);
});

test("ModelAPI.create", async () => {
  const row = await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Jim",
    married: false,
    favouriteNumber: 10,
  });
  expect(row.name).toEqual("Jim");
  expect(row.married).toEqual(false);
  expect(row.favouriteNumber).toEqual(10);
  expect(KSUID.parse(row.id).string).toEqual(row.id);
});

test("ModelAPI.create - non-ksuid id", async () => {
  const row = await personAPI.create({
    id: "not-a-ksuid",
    name: "Jim",
    married: false,
    favouriteNumber: 10,
  });
  expect(row.name).toEqual("Jim");
  expect(row.married).toEqual(false);
  expect(row.favouriteNumber).toEqual(10);
  expect(row.id).toEqual(row.id);
});

test("ModelAPI.create - throws if not not null constraint violation", async () => {
  await expect(
    authorAPI.create({
      id: KSUID.randomSync().string,
      name: null,
    })
  ).rejects.toThrow(
    'null value in column "name" of relation "author" violates not-null constraint'
  );
});

test("ModelAPI.create - throws if database constraint fails", async () => {
  const row = await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Jim",
    married: false,
    favouriteNumber: 10,
  });
  const promise = personAPI.create({
    id: row.id,
    name: "Jim",
    married: false,
    favouriteNumber: 10,
  });
  await expect(promise).rejects.toThrow(
    `duplicate key value violates unique constraint "person_pkey"`
  );
});

test("ModelAPI.create - arrays", async () => {
  const person = await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Jim",
    married: false,
    favouriteNumber: 10,
  });

  const row = await postAPI.create({
    id: "id",
    title: "My Post",
    tags: ["tag 1", "tag 2"],
    rating: 1.23,
    authorId: person.id,
  });
  expect(row.tags).toEqual(["tag 1", "tag 2"]);
  expect(row.rating).toEqual(1.23);
});

test("ModelAPI.findOne", async () => {
  const created = await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Jim",
    married: false,
    favouriteNumber: 10,
  });
  const row = await personAPI.findOne({
    id: created.id,
  });
  expect(row).toEqual(created);
});

test("ModelAPI.findOne - relationships - one to many", async () => {
  const person = await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Jim",
    married: false,
    favouriteNumber: 10,
  });
  const post = await postAPI.create({
    id: KSUID.randomSync().string,
    title: "My Post",
    authorId: person.id,
  });
  const row = await personAPI.findOne({
    posts: {
      id: post.id,
    },
  });
  expect(row.name).toEqual("Jim");
  expect(row.id).toEqual(person.id);
});

test("ModelAPI.findOne - return null if not found", async () => {
  const row = await personAPI.findOne({
    id: "doesntexist",
  });
  expect(row).toEqual(null);
});

test("ModelAPI.findMany", async () => {
  await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Jim",
    married: false,
    favouriteNumber: 10,
  });
  const bob = await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Bob",
    married: true,
    favouriteNumber: 11,
  });
  const sally = await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Sally",
    married: true,
    favouriteNumber: 12,
  });
  const rows = await personAPI.findMany({
    where: {
      married: true,
    },
  });
  expect(rows.length).toEqual(2);
  expect(rows.map((x) => x.id).sort()).toEqual([bob.id, sally.id].sort());
});

test("ModelAPI.findMany - no where conditions", async () => {
  await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Jim",
  });
  await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Bob",
  });

  const rows = await personAPI.findMany({});

  expect(rows.length).toEqual(2);
});

test("ModelAPI.findMany - startsWith", async () => {
  const jim = await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Jim",
  });
  await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Bob",
  });
  const rows = await personAPI.findMany({
    where: {
      name: {
        startsWith: "Ji",
      },
    },
  });
  expect(rows.length).toEqual(1);
  expect(rows[0].id).toEqual(jim.id);
});

test("ModelAPI.findMany - endsWith", async () => {
  const jim = await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Jim",
  });
  await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Bob",
  });
  const rows = await personAPI.findMany({
    where: {
      name: {
        endsWith: "im",
      },
    },
  });
  expect(rows.length).toEqual(1);
  expect(rows[0].id).toEqual(jim.id);
});

test("ModelAPI.findMany - contains", async () => {
  const billy = await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Billy",
  });
  const sally = await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Sally",
  });
  await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Jim",
  });
  const rows = await personAPI.findMany({
    where: {
      name: {
        contains: "ll",
      },
    },
  });
  expect(rows.length).toEqual(2);
  expect(rows.map((x) => x.id).sort()).toEqual([billy.id, sally.id].sort());
});

test("ModelAPI.findMany - oneOf", async () => {
  const billy = await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Billy",
  });
  const sally = await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Sally",
  });
  await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Jim",
  });
  const rows = await personAPI.findMany({
    where: {
      name: {
        oneOf: ["Billy", "Sally"],
      },
    },
  });
  expect(rows.length).toEqual(2);
  expect(rows.map((x) => x.id).sort()).toEqual([billy.id, sally.id].sort());
});

test("ModelAPI.findMany - notEquals on id", async () => {
  const p1 = await personAPI.create({
    id: KSUID.randomSync().string,
    favouriteNumber: 1,
  });
  const p2 = await personAPI.create({
    id: KSUID.randomSync().string,
    favouriteNumber: 2,
  });
  const rows = await personAPI.findMany({
    where: {
      id: {
        notEquals: p1.id,
      },
    },
  });
  expect(rows.length).toEqual(1);
  expect(rows[0].id).toEqual(p2.id);
});

test("ModelAPI.findMany - greaterThan", async () => {
  await personAPI.create({
    id: KSUID.randomSync().string,
    favouriteNumber: 1,
  });
  const p = await personAPI.create({
    id: KSUID.randomSync().string,
    favouriteNumber: 2,
  });
  const rows = await personAPI.findMany({
    where: {
      favouriteNumber: {
        greaterThan: 1,
      },
    },
  });
  expect(rows.length).toEqual(1);
  expect(rows[0].id).toEqual(p.id);
});

test("ModelAPI.findMany - greaterThanOrEquals", async () => {
  await personAPI.create({
    id: KSUID.randomSync().string,
    favouriteNumber: 1,
  });
  const p = await personAPI.create({
    id: KSUID.randomSync().string,
    favouriteNumber: 2,
  });
  const p2 = await personAPI.create({
    id: KSUID.randomSync().string,
    favouriteNumber: 3,
  });
  const rows = await personAPI.findMany({
    where: {
      favouriteNumber: {
        greaterThanOrEquals: 2,
      },
    },
  });
  expect(rows.length).toEqual(2);
  expect(rows.map((x) => x.id).sort()).toEqual([p.id, p2.id].sort());
});

test("ModelAPI.findMany - lessThan", async () => {
  const p = await personAPI.create({
    id: KSUID.randomSync().string,
    favouriteNumber: 1,
  });
  await personAPI.create({
    id: KSUID.randomSync().string,
    favouriteNumber: 2,
  });
  const rows = await personAPI.findMany({
    where: {
      favouriteNumber: {
        lessThan: 2,
      },
    },
  });
  expect(rows.length).toEqual(1);
  expect(rows[0].id).toEqual(p.id);
});

test("ModelAPI.findMany - lessThanOrEquals", async () => {
  const p = await personAPI.create({
    id: KSUID.randomSync().string,
    favouriteNumber: 1,
  });
  const p2 = await personAPI.create({
    id: KSUID.randomSync().string,
    favouriteNumber: 2,
  });
  await personAPI.create({
    id: KSUID.randomSync().string,
    favouriteNumber: 3,
  });
  const rows = await personAPI.findMany({
    where: {
      favouriteNumber: {
        lessThanOrEquals: 2,
      },
    },
  });
  expect(rows.length).toEqual(2);
  expect(rows.map((x) => x.id).sort()).toEqual([p.id, p2.id].sort());
});

test("ModelAPI.findMany - before", async () => {
  const p = await personAPI.create({
    id: KSUID.randomSync().string,
    date: new Date("2022-01-01"),
  });
  await personAPI.create({
    id: KSUID.randomSync().string,
    date: new Date("2022-01-02"),
  });
  const rows = await personAPI.findMany({
    where: {
      date: {
        before: new Date("2022-01-02"),
      },
    },
  });
  expect(rows.length).toEqual(1);
  expect(rows[0].id).toEqual(p.id);
});

test("ModelAPI.findMany - empty where", async () => {
  const p = await personAPI.create({
    id: KSUID.randomSync().string,
    date: new Date("2022-01-01"),
  });

  const p2 = await personAPI.create({
    id: KSUID.randomSync().string,
    date: new Date("2022-01-02"),
  });

  // with no param specified at all
  const rows = await personAPI.findMany();

  expect(rows.map((r) => r.id).sort()).toEqual([p, p2].map((r) => r.id).sort());

  // with empty object
  const rows2 = await personAPI.findMany({});

  expect(rows2.map((r) => r.id).sort()).toEqual(
    [p, p2].map((r) => r.id).sort()
  );
});

test("ModelAPI.findMany - onOrBefore", async () => {
  const p = await personAPI.create({
    id: KSUID.randomSync().string,
    date: new Date("2022-01-01"),
  });
  const p2 = await personAPI.create({
    id: KSUID.randomSync().string,
    date: new Date("2022-01-02"),
  });
  await personAPI.create({
    id: KSUID.randomSync().string,
    date: new Date("2022-01-03"),
  });
  const rows = await personAPI.findMany({
    where: {
      date: {
        onOrBefore: new Date("2022-01-02"),
      },
    },
  });
  expect(rows.length).toEqual(2);
  expect(rows.map((x) => x.id).sort()).toEqual([p.id, p2.id].sort());
});

test("ModelAPI.findMany - limit", async () => {
  await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Jim",
    married: false,
    favouriteNumber: 10,
  });
  await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Bob",
    married: true,
    favouriteNumber: 11,
  });
  await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Sally",
    married: true,
    favouriteNumber: 12,
  });

  const rows = await personAPI.findMany({
    limit: 2,
    orderBy: {
      favouriteNumber: "asc",
    },
  });

  expect(rows.map((r) => r.name)).toEqual(["Jim", "Bob"]);
});

test("ModelAPI.findMany - orderBy", async () => {
  await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Jim",
    married: false,
    favouriteNumber: 10,
    date: new Date(2023, 12, 29),
  });
  await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Bob",
    married: true,
    favouriteNumber: 11,
    date: new Date(2023, 12, 30),
  });
  await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Sally",
    married: true,
    favouriteNumber: 12,
    date: new Date(2023, 12, 31),
  });

  const ascendingNames = await personAPI.findMany({
    orderBy: {
      name: "asc",
    },
  });

  expect(ascendingNames.map((r) => r.name)).toEqual(["Bob", "Jim", "Sally"]);

  const descendingNames = await personAPI.findMany({
    orderBy: {
      name: "desc",
    },
  });

  expect(descendingNames.map((r) => r.name)).toEqual(["Sally", "Jim", "Bob"]);

  const ascendingFavouriteNumbers = await personAPI.findMany({
    orderBy: {
      favouriteNumber: "asc",
    },
  });

  expect(ascendingFavouriteNumbers.map((r) => r.name)).toEqual([
    "Jim",
    "Bob",
    "Sally",
  ]);

  const descendingDates = await personAPI.findMany({
    orderBy: {
      date: "desc",
    },
  });

  expect(descendingDates.map((r) => r.name)).toEqual(["Sally", "Bob", "Jim"]);
});

test("ModelAPI.findMany - orderBy ASC and DESC capitalised", async () => {
  await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Jim",
    married: false,
    favouriteNumber: 10,
    date: new Date(2023, 12, 29),
  });
  await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Bob",
    married: true,
    favouriteNumber: 11,
    date: new Date(2023, 12, 30),
  });
  await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Sally",
    married: true,
    favouriteNumber: 12,
    date: new Date(2023, 12, 31),
  });

  const ascendingNames = await personAPI.findMany({
    orderBy: {
      name: "ASC",
    },
  });

  expect(ascendingNames.map((r) => r.name)).toEqual(["Bob", "Jim", "Sally"]);

  const descendingNames = await personAPI.findMany({
    orderBy: {
      name: "DESC",
    },
  });

  expect(descendingNames.map((r) => r.name)).toEqual(["Sally", "Jim", "Bob"]);
});

test("ModelAPI.findMany - offset", async () => {
  await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Jim",
    married: false,
    favouriteNumber: 10,
    date: new Date(2023, 12, 29),
  });
  await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Bob",
    married: true,
    favouriteNumber: 11,
    date: new Date(2023, 12, 30),
  });
  await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Sally",
    married: true,
    favouriteNumber: 12,
    date: new Date(2023, 12, 31),
  });

  const rows = await personAPI.findMany({
    offset: 1,
    limit: 2,
    orderBy: {
      name: "asc",
    },
  });

  expect(rows.map((r) => r.name)).toEqual(["Jim", "Sally"]);

  const rows2 = await personAPI.findMany({
    offset: 2,
    orderBy: {
      name: "asc",
    },
  });

  expect(rows2.map((r) => r.name)).toEqual(["Sally"]);

  const rows3 = await personAPI.findMany({
    offset: 1,
    orderBy: {
      name: "asc",
    },
    limit: 1,
  });

  expect(rows3.map((r) => r.name)).toEqual(["Jim"]);
});

test("ModelAPI.findMany - after", async () => {
  await personAPI.create({
    id: KSUID.randomSync().string,
    date: new Date("2022-01-01"),
  });
  const p = await personAPI.create({
    id: KSUID.randomSync().string,
    date: new Date("2022-01-02"),
  });
  const rows = await personAPI.findMany({
    where: {
      date: {
        after: new Date("2022-01-01"),
      },
    },
  });
  expect(rows.length).toEqual(1);
  expect(rows[0].id).toEqual(p.id);
});

test("ModelAPI.findMany - onOrAfter", async () => {
  await personAPI.create({
    id: KSUID.randomSync().string,
    date: new Date("2022-01-01"),
  });
  const p = await personAPI.create({
    id: KSUID.randomSync().string,
    date: new Date("2022-01-02"),
  });
  const p2 = await personAPI.create({
    id: KSUID.randomSync().string,
    date: new Date("2022-01-03"),
  });
  const rows = await personAPI.findMany({
    where: {
      date: {
        onOrAfter: new Date("2022-01-02"),
      },
    },
  });
  expect(rows.length).toEqual(2);
  expect(rows.map((x) => x.id).sort()).toEqual([p.id, p2.id].sort());
});

test("ModelAPI.findMany - equals", async () => {
  const p = await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Jim",
  });
  await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Sally",
  });
  const rows = await personAPI.findMany({
    where: {
      name: {
        equals: "Jim",
      },
    },
  });
  expect(rows.length).toEqual(1);
  expect(rows[0].id).toEqual(p.id);
});

test("ModelAPI.findMany - notEquals", async () => {
  const p = await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Jim",
  });
  await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Sally",
  });
  const rows = await personAPI.findMany({
    where: {
      name: {
        notEquals: "Sally",
      },
    },
  });
  expect(rows.length).toEqual(1);
  expect(rows[0].id).toEqual(p.id);
});

test("ModelAPI.findMany - relationships - one to many", async () => {
  const person = await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Jim",
  });
  const person2 = await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Bob",
  });
  const post1 = await postAPI.create({
    id: KSUID.randomSync().string,
    title: "My First Post",
    authorId: person.id,
  });
  const post2 = await postAPI.create({
    id: KSUID.randomSync().string,
    title: "My Second Post",
    authorId: person.id,
  });
  await postAPI.create({
    id: KSUID.randomSync().string,
    title: "My Third Post",
    authorId: person2.id,
  });

  const posts = await postAPI.findMany({
    where: {
      author: {
        name: "Jim",
      },
    },
  });
  expect(posts.length).toEqual(2);
  expect(posts.map((x) => x.id).sort()).toEqual([post1.id, post2.id].sort());
});

test("ModelAPI.findMany - relationships - many to one", async () => {
  const person = await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Jim",
  });
  await postAPI.create({
    id: KSUID.randomSync().string,
    title: "My First Post",
    authorId: person.id,
  });
  await postAPI.create({
    id: KSUID.randomSync().string,
    title: "My Second Post",
    authorId: person.id,
  });
  await postAPI.create({
    id: KSUID.randomSync().string,
    title: "My Second Post",
    authorId: person.id,
  });

  const people = await personAPI.findMany({
    where: {
      posts: {
        title: {
          startsWith: "My ",
          endsWith: " Post",
        },
      },
    },
  });

  // This tests that many to one joins work for findMany() but also
  // that the same row is not returned more than once e.g. Jim has
  // three posts but should only be returned once
  expect(people.length).toEqual(1);
  expect(people[0].id).toEqual(person.id);
});

test("ModelAPI.findMany - relationships - duplicate joins handled", async () => {
  const person = await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Jim",
  });
  const person2 = await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Bob",
  });
  const post1 = await postAPI.create({
    id: KSUID.randomSync().string,
    title: "My First Post",
    authorId: person.id,
  });
  const post2 = await postAPI.create({
    id: KSUID.randomSync().string,
    title: "My Second Post",
    authorId: person2.id,
  });

  const posts = await postAPI
    .where({
      author: {
        name: "Jim",
      },
    })
    .findMany();

  expect(posts.length).toEqual(1);
  expect(posts.map((x) => x.id).sort()).toEqual([post1.id].sort());
});

test("ModelAPI.update", async () => {
  let jim = await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Jim",
    married: false,
    favouriteNumber: 10,
  });
  let bob = await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Bob",
    married: false,
    favouriteNumber: 11,
  });
  jim = await personAPI.update(
    {
      id: jim.id,
    },
    {
      married: true,
    }
  );
  expect(jim.married).toEqual(true);
  expect(jim.name).toEqual("Jim");

  bob = await personAPI.findOne({ id: bob.id });
  expect(bob.married).toEqual(false);
});

test("ModelAPI.update - throws if not found", async () => {
  const result = personAPI.update(
    {
      id: "doesntexist",
    },
    {
      married: true,
    }
  );
  await expect(result).rejects.toThrow("no result");
});

test("ModelAPI.update - throws if not not null constraint violation", async () => {
  const jim = await authorAPI.create({
    id: KSUID.randomSync().string,
    name: "jim",
  });

  const result = authorAPI.update(
    {
      id: jim.id,
    },
    {
      name: null,
    }
  );

  await expect(result).rejects.toThrow(
    'null value in column "name" of relation "author" violates not-null constraint'
  );
});

test("ModelAPI.delete", async () => {
  const jim = await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Jim",
  });
  const id = jim.id;
  const deletedId = await personAPI.delete({
    name: "Jim",
  });

  expect(deletedId).toEqual(id);
  await expect(personAPI.findOne({ id })).resolves.toEqual(null);
});

describe("QueryBuilder", () => {
  test("ModelAPI chained findMany with offset/limit/order by", async () => {
    await postAPI.create({
      id: KSUID.randomSync().string,
      title: "adam",
    });
    await postAPI.create({
      id: KSUID.randomSync().string,
      title: "dave",
    });
    const three = await postAPI.create({
      id: KSUID.randomSync().string,
      title: "jon",
    });
    const four = await postAPI.create({
      id: KSUID.randomSync().string,
      title: "jon bretman",
    });

    const results = await postAPI
      .where({ title: { startsWith: "jon" } })
      .findMany({
        limit: 1,
        offset: 1,
        orderBy: {
          title: "asc",
        },
      });

    // because we've offset by 1, adam should not appear in the results even though
    // the query constraints match adam
    expect(results).toEqual([four]);
  });

  test("ModelAPI.findMany - complex query", async () => {
    const p = await personAPI.create({
      id: KSUID.randomSync().string,
      name: "Jake",
      favouriteNumber: 8,
      date: new Date("2021-12-31"),
    });
    await personAPI.create({
      id: KSUID.randomSync().string,
      name: "Jane",
      favouriteNumber: 12,
      date: new Date("2022-01-11"),
    });
    const p2 = await personAPI.create({
      id: KSUID.randomSync().string,
      name: "Billy",
      favouriteNumber: 16,
      date: new Date("2022-01-05"),
    });

    const rows = await personAPI
      // Will match Jane
      .where({
        name: {
          startsWith: "J",
          endsWith: "e",
        },
        favouriteNumber: {
          lessThan: 10,
        },
      })
      .findMany();
    expect(rows.length).toEqual(1);
    expect(rows.map((x) => x.id).sort()).toEqual([p.id].sort());
  });

  test("ModelAPI chained delete", async () => {
    const p = await personAPI.create({
      id: KSUID.randomSync().string,
      name: "Jake",
      favouriteNumber: 8,
      date: new Date("2021-12-31"),
    });

    const deletedId = await personAPI.where({ id: p.id }).delete();

    expect(deletedId).toEqual(p.id);
  });

  test("ModelAPI chained delete - non existent id", async () => {
    const fakeId = "xxx";

    // the error message returned from the runtime will actually be 'record not found'
    // but this is handled at handleRequest level
    // no result is the error msg returned by kysely.
    await expect(personAPI.where({ id: fakeId }).delete()).rejects.toThrow(
      "no result"
    );
  });

  test("ModelAPI chained findOne", async () => {
    const p = await personAPI.create({
      id: KSUID.randomSync().string,
      name: "Jake",
      favouriteNumber: 8,
      date: new Date("2021-12-31"),
    });

    const jake = await personAPI.where({ id: p.id }).findOne();

    expect(jake).toEqual(p);
  });

  test("ModelAPI chained update", async () => {
    const p1 = await postAPI.create({
      id: KSUID.randomSync().string,
      title: "adam",
    });
    const p2 = await postAPI.create({
      id: KSUID.randomSync().string,
      title: "adam",
    });
    const p3 = await postAPI.create({
      id: KSUID.randomSync().string,
      title: "adam",
    });

    const updatedRow = await postAPI
      .where({ id: p2.id })
      .update({ title: "adam 2" });

    expect(updatedRow.title).toEqual("adam 2");
    expect(updatedRow.id).toEqual(p2.id);

    // will fail because there is more than 1 row matching the constraints (p1 and p3)
    await expect(
      postAPI
        .where({
          title: "adam",
        })
        .update({ title: "bob" })
    ).rejects.toThrowError(
      "more than one row matched update constraints - only unique fields should be used when updating."
    );

    // will fail because there are no rows to update
    await expect(
      postAPI
        .where({
          title: "no match",
        })
        .update({ title: "bob" })
    ).resolves.toEqual(null);
  });

  test("ModelAPI.findMany - array equals", async () => {
    const person = await personAPI.create({
      id: KSUID.randomSync().string,
      name: "Jake",
      favouriteNumber: 8,
      date: new Date("2021-12-31"),
    });
    const p1 = await postAPI.create({
      id: KSUID.randomSync().string,
      title: "Post 1",
      tags: ["Tag 1", "Tag 2"],
      authorId: person.id,
    });
    const p2 = await postAPI.create({
      id: KSUID.randomSync().string,
      title: "Post 2",
      tags: ["Tag 2", "Tag 3"],
      authorId: person.id,
    });
    const p3 = await postAPI.create({
      id: KSUID.randomSync().string,
      title: "Post 2",
      tags: [],
      authorId: person.id,
    });
    const p4 = await postAPI.create({
      id: KSUID.randomSync().string,
      title: "Post 2",
      tags: null,
      authorId: person.id,
    });

    const rows = await postAPI.findMany({
      where: {
        tags: {
          equals: ["Tag 1", "Tag 2"],
        },
      },
    });

    expect(rows.length).toEqual(1);
    expect(rows.map((x) => x.id).sort()).toEqual([p1.id].sort());
  });

  test("ModelAPI.findMany - array equals implicit", async () => {
    const person = await personAPI.create({
      id: KSUID.randomSync().string,
      name: "Jake",
      favouriteNumber: 8,
      date: new Date("2021-12-31"),
    });
    const p1 = await postAPI.create({
      id: KSUID.randomSync().string,
      title: "Post 1",
      tags: ["Tag 1", "Tag 2"],
      authorId: person.id,
    });
    const p2 = await postAPI.create({
      id: KSUID.randomSync().string,
      title: "Post 2",
      tags: ["Tag 2", "Tag 3"],
      authorId: person.id,
    });
    const p3 = await postAPI.create({
      id: KSUID.randomSync().string,
      title: "Post 2",
      tags: [],
      authorId: person.id,
    });

    const rows = await postAPI.findMany({
      where: {
        tags: ["Tag 1", "Tag 2"],
      },
    });

    expect(rows.length).toEqual(1);
    expect(rows.map((x) => x.id).sort()).toEqual([p1.id].sort());
  });

  test("ModelAPI.findMany - array not equals", async () => {
    const person = await personAPI.create({
      id: KSUID.randomSync().string,
      name: "Jake",
      favouriteNumber: 8,
      date: new Date("2021-12-31"),
    });
    const p1 = await postAPI.create({
      id: KSUID.randomSync().string,
      title: "Post 1",
      tags: ["Tag 1", "Tag 2"],
      authorId: person.id,
    });
    const p2 = await postAPI.create({
      id: KSUID.randomSync().string,
      title: "Post 2",
      tags: ["Tag 2", "Tag 3"],
      authorId: person.id,
    });
    const p3 = await postAPI.create({
      id: KSUID.randomSync().string,
      title: "Post 2",
      tags: [],
      authorId: person.id,
    });
    const p4 = await postAPI.create({
      id: KSUID.randomSync().string,
      title: "Post 2",
      tags: null,
      authorId: person.id,
    });

    const rows = await postAPI.findMany({
      where: {
        tags: {
          notEquals: ["Tag 1", "Tag 2"],
        },
      },
    });

    expect(rows.length).toEqual(3);
    expect(rows.map((x) => x.id).sort()).toEqual([p4.id, p3.id, p2.id].sort());
  });

  test("ModelAPI.findMany - array any equals", async () => {
    const person = await personAPI.create({
      id: KSUID.randomSync().string,
      name: "Jake",
      favouriteNumber: 8,
      date: new Date("2021-12-31"),
    });
    const p1 = await postAPI.create({
      id: KSUID.randomSync().string,
      title: "Post 1",
      tags: ["Tag 1", "Tag 2"],
      authorId: person.id,
    });
    const p2 = await postAPI.create({
      id: KSUID.randomSync().string,
      title: "Post 2",
      tags: ["Tag 2", "Tag 3"],
      authorId: person.id,
    });
    const p3 = await postAPI.create({
      id: KSUID.randomSync().string,
      title: "Post 2",
      tags: [],
      authorId: person.id,
    });
    const p4 = await postAPI.create({
      id: KSUID.randomSync().string,
      title: "Post 2",
      tags: null,
      authorId: person.id,
    });
    const p5 = await postAPI.create({
      id: KSUID.randomSync().string,
      title: "Post 2",
      tags: ["Tag 3"],
      authorId: person.id,
    });

    const rows = await postAPI.findMany({
      where: {
        tags: {
          any: {
            equals: "Tag 2",
          },
        },
      },
    });

    expect(rows.length).toEqual(2);
    expect(rows.map((x) => x.id).sort()).toEqual([p1.id, p2.id].sort());
  });

  test("ModelAPI.findMany - array any not equals", async () => {
    const person = await personAPI.create({
      id: KSUID.randomSync().string,
      name: "Jake",
      favouriteNumber: 8,
      date: new Date("2021-12-31"),
    });
    const p1 = await postAPI.create({
      id: KSUID.randomSync().string,
      title: "Post 1",
      tags: ["Tag 1", "Tag 2"],
      authorId: person.id,
    });
    const p2 = await postAPI.create({
      id: KSUID.randomSync().string,
      title: "Post 2",
      tags: ["Tag 2", "Tag 3"],
      authorId: person.id,
    });
    const p3 = await postAPI.create({
      id: KSUID.randomSync().string,
      title: "Post 2",
      tags: [],
      authorId: person.id,
    });
    const p4 = await postAPI.create({
      id: KSUID.randomSync().string,
      title: "Post 2",
      tags: null,
      authorId: person.id,
    });
    const p5 = await postAPI.create({
      id: KSUID.randomSync().string,
      title: "Post 2",
      tags: ["Tag 3"],
      authorId: person.id,
    });

    const rows = await postAPI.findMany({
      where: {
        tags: {
          any: {
            notEquals: "Tag 3",
          },
        },
      },
    });

    expect(rows.length).toEqual(2);
    expect(rows.map((x) => x.id).sort()).toEqual([p1.id, p3.id].sort());
  });

  test("ModelAPI.findMany - array all equals", async () => {
    const person = await personAPI.create({
      id: KSUID.randomSync().string,
      name: "Jake",
      favouriteNumber: 8,
      date: new Date("2021-12-31"),
    });
    const p1 = await postAPI.create({
      id: KSUID.randomSync().string,
      title: "Post 1",
      tags: ["Tag 1", "Tag 2"],
      authorId: person.id,
    });
    const p2 = await postAPI.create({
      id: KSUID.randomSync().string,
      title: "Post 2",
      tags: ["Tag 2", "Tag 3"],
      authorId: person.id,
    });
    const p3 = await postAPI.create({
      id: KSUID.randomSync().string,
      title: "Post 2",
      tags: [],
      authorId: person.id,
    });
    const p4 = await postAPI.create({
      id: KSUID.randomSync().string,
      title: "Post 2",
      tags: null,
      authorId: person.id,
    });
    const p5 = await postAPI.create({
      id: KSUID.randomSync().string,
      title: "Post 2",
      tags: ["Tag 2", "Tag 2"],
      authorId: person.id,
    });

    const rows = await postAPI.findMany({
      where: {
        tags: {
          all: {
            equals: "Tag 2",
          },
        },
      },
    });

    expect(rows.length).toEqual(2);
    expect(rows.map((x) => x.id).sort()).toEqual([p3.id, p5.id].sort());
  });

  test("ModelAPI.findMany - array all not equals", async () => {
    const person = await personAPI.create({
      id: KSUID.randomSync().string,
      name: "Jake",
      favouriteNumber: 8,
      date: new Date("2021-12-31"),
    });
    const p1 = await postAPI.create({
      id: KSUID.randomSync().string,
      title: "Post 1",
      tags: ["Tag 1", "Tag 2"],
      authorId: person.id,
    });
    const p2 = await postAPI.create({
      id: KSUID.randomSync().string,
      title: "Post 2",
      tags: ["Tag 2", "Tag 3"],
      authorId: person.id,
    });
    const p3 = await postAPI.create({
      id: KSUID.randomSync().string,
      title: "Post 2",
      tags: [],
      authorId: person.id,
    });
    const p4 = await postAPI.create({
      id: KSUID.randomSync().string,
      title: "Post 2",
      tags: null,
      authorId: person.id,
    });
    const p5 = await postAPI.create({
      id: KSUID.randomSync().string,
      title: "Post 2",
      tags: ["Tag 2", "Tag 2"],
      authorId: person.id,
    });

    const rows = await postAPI.findMany({
      where: {
        tags: {
          all: {
            notEquals: "Tag 2",
          },
        },
      },
    });

    expect(rows.length).toEqual(2);
    expect(rows.map((x) => x.id).sort()).toEqual([p1.id, p2.id].sort());
  });
});
