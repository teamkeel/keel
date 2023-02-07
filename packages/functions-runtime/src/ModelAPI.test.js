import { test, expect, beforeEach } from "vitest";
const { ModelAPI, DatabaseError } = require("./ModelAPI");
const { sql } = require("kysely");
const { getDatabase } = require("./database");
const KSUID = require("ksuid");

process.env.DB_CONN_TYPE = "pg";
process.env.DB_CONN = `postgresql://postgres:postgres@localhost:5432/functions-runtime`;

let personAPI;
let postAPI;
let authorAPI;

beforeEach(async () => {
  const db = getDatabase();

  await sql`
  DROP TABLE IF EXISTS post;
  DROP TABLE IF EXISTS person;
  DROP TABLE IF EXISTS author;

  CREATE TABLE person(
      id               text PRIMARY KEY,
      name             text UNIQUE,
      married          boolean,
      favourite_number integer,
      date             timestamp
  );
  CREATE TABLE post(
    id               text PRIMARY KEY,
    title            text,
    author_id        text references person(id)
  );
  CREATE TABLE author(
    id               text PRIMARY KEY,
    name             text NOT NULL
  );
  `.execute(db);

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

  personAPI = new ModelAPI(
    "person",
    () => {
      return {
        id: KSUID.randomSync().string,
        date: new Date("2022-01-01"),
      };
    },
    db,
    tableConfigMap
  );

  postAPI = new ModelAPI(
    "post",
    () => {
      return {
        id: KSUID.randomSync().string,
      };
    },
    db,
    tableConfigMap
  );

  authorAPI = new ModelAPI(
    "author",
    () => {
      return {
        id: KSUID.randomSync().string,
      };
    },
    db,
    tableConfigMap
  );
});

test("ModelAPI.create", async () => {
  const row = await personAPI.create({
    name: "Jim",
    married: false,
    favouriteNumber: 10,
  });
  expect(row.name).toEqual("Jim");
  expect(row.married).toEqual(false);
  expect(row.date).toEqual(new Date("2022-01-01"));
  expect(row.favouriteNumber).toEqual(10);
  expect(KSUID.parse(row.id).string).toEqual(row.id);
});

test("ModelAPI.create - throws if not not null constraint violation", async () => {
  await expect(
    authorAPI.create({
      name: null,
    })
  ).rejects.toThrow('null value in column "name" violates not-null constraint');
});

test("ModelAPI.create - throws if database constraint fails", async () => {
  const row = await personAPI.create({
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

test("ModelAPI.findOne", async () => {
  const created = await personAPI.create({
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
    name: "Jim",
    married: false,
    favouriteNumber: 10,
  });
  const post = await postAPI.create({
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
  const jim = await personAPI.create({
    name: "Jim",
    married: false,
    favouriteNumber: 10,
  });
  const bob = await personAPI.create({
    name: "Bob",
    married: true,
    favouriteNumber: 11,
  });
  const sally = await personAPI.create({
    name: "Sally",
    married: true,
    favouriteNumber: 12,
  });
  const rows = await personAPI.findMany({
    married: true,
  });
  expect(rows.length).toEqual(2);
  expect(rows.map((x) => x.id).sort()).toEqual([bob.id, sally.id].sort());
});

test("ModelAPI.findMany - no where conditions", async () => {
  const jim = await personAPI.create({
    name: "Jim",
  });
  await personAPI.create({
    name: "Bob",
  });

  const rows = await personAPI.findMany({});

  expect(rows.length).toEqual(2);
});

test("ModelAPI.findMany - startsWith", async () => {
  const jim = await personAPI.create({
    name: "Jim",
  });
  await personAPI.create({
    name: "Bob",
  });
  const rows = await personAPI.findMany({
    name: {
      startsWith: "Ji",
    },
  });
  expect(rows.length).toEqual(1);
  expect(rows[0].id).toEqual(jim.id);
});

test("ModelAPI.findMany - endsWith", async () => {
  const jim = await personAPI.create({
    name: "Jim",
  });
  await personAPI.create({
    name: "Bob",
  });
  const rows = await personAPI.findMany({
    name: {
      endsWith: "im",
    },
  });
  expect(rows.length).toEqual(1);
  expect(rows[0].id).toEqual(jim.id);
});

test("ModelAPI.findMany - contains", async () => {
  const billy = await personAPI.create({
    name: "Billy",
  });
  const sally = await personAPI.create({
    name: "Sally",
  });
  await personAPI.create({
    name: "Jim",
  });
  const rows = await personAPI.findMany({
    name: {
      contains: "ll",
    },
  });
  expect(rows.length).toEqual(2);
  expect(rows.map((x) => x.id).sort()).toEqual([billy.id, sally.id].sort());
});

test("ModelAPI.findMany - oneOf", async () => {
  const billy = await personAPI.create({
    name: "Billy",
  });
  const sally = await personAPI.create({
    name: "Sally",
  });
  await personAPI.create({
    name: "Jim",
  });
  const rows = await personAPI.findMany({
    name: {
      oneOf: ["Billy", "Sally"],
    },
  });
  expect(rows.length).toEqual(2);
  expect(rows.map((x) => x.id).sort()).toEqual([billy.id, sally.id].sort());
});

test("ModelAPI.findMany - greaterThan", async () => {
  await personAPI.create({
    favouriteNumber: 1,
  });
  const p = await personAPI.create({
    favouriteNumber: 2,
  });
  const rows = await personAPI.findMany({
    favouriteNumber: {
      greaterThan: 1,
    },
  });
  expect(rows.length).toEqual(1);
  expect(rows[0].id).toEqual(p.id);
});

test("ModelAPI.findMany - greaterThanOrEquals", async () => {
  await personAPI.create({
    favouriteNumber: 1,
  });
  const p = await personAPI.create({
    favouriteNumber: 2,
  });
  const p2 = await personAPI.create({
    favouriteNumber: 3,
  });
  const rows = await personAPI.findMany({
    favouriteNumber: {
      greaterThanOrEquals: 2,
    },
  });
  expect(rows.length).toEqual(2);
  expect(rows.map((x) => x.id).sort()).toEqual([p.id, p2.id].sort());
});

test("ModelAPI.findMany - lessThan", async () => {
  const p = await personAPI.create({
    favouriteNumber: 1,
  });
  await personAPI.create({
    favouriteNumber: 2,
  });
  const rows = await personAPI.findMany({
    favouriteNumber: {
      lessThan: 2,
    },
  });
  expect(rows.length).toEqual(1);
  expect(rows[0].id).toEqual(p.id);
});

test("ModelAPI.findMany - lessThanOrEquals", async () => {
  const p = await personAPI.create({
    favouriteNumber: 1,
  });
  const p2 = await personAPI.create({
    favouriteNumber: 2,
  });
  await personAPI.create({
    favouriteNumber: 3,
  });
  const rows = await personAPI.findMany({
    favouriteNumber: {
      lessThanOrEquals: 2,
    },
  });
  expect(rows.length).toEqual(2);
  expect(rows.map((x) => x.id).sort()).toEqual([p.id, p2.id].sort());
});

test("ModelAPI.findMany - before", async () => {
  const p = await personAPI.create({
    date: new Date("2022-01-01"),
  });
  await personAPI.create({
    date: new Date("2022-01-02"),
  });
  const rows = await personAPI.findMany({
    date: {
      before: new Date("2022-01-02"),
    },
  });
  expect(rows.length).toEqual(1);
  expect(rows[0].id).toEqual(p.id);
});

test("ModelAPI.findMany - onOrBefore", async () => {
  const p = await personAPI.create({
    date: new Date("2022-01-01"),
  });
  const p2 = await personAPI.create({
    date: new Date("2022-01-02"),
  });
  await personAPI.create({
    date: new Date("2022-01-03"),
  });
  const rows = await personAPI.findMany({
    date: {
      onOrBefore: new Date("2022-01-02"),
    },
  });
  expect(rows.length).toEqual(2);
  expect(rows.map((x) => x.id).sort()).toEqual([p.id, p2.id].sort());
});

test("ModelAPI.findMany - after", async () => {
  await personAPI.create({
    date: new Date("2022-01-01"),
  });
  const p = await personAPI.create({
    date: new Date("2022-01-02"),
  });
  const rows = await personAPI.findMany({
    date: {
      after: new Date("2022-01-01"),
    },
  });
  expect(rows.length).toEqual(1);
  expect(rows[0].id).toEqual(p.id);
});

test("ModelAPI.findMany - onOrAfter", async () => {
  await personAPI.create({
    date: new Date("2022-01-01"),
  });
  const p = await personAPI.create({
    date: new Date("2022-01-02"),
  });
  const p2 = await personAPI.create({
    date: new Date("2022-01-03"),
  });
  const rows = await personAPI.findMany({
    date: {
      onOrAfter: new Date("2022-01-02"),
    },
  });
  expect(rows.length).toEqual(2);
  expect(rows.map((x) => x.id).sort()).toEqual([p.id, p2.id].sort());
});

test("ModelAPI.findMany - equals", async () => {
  const p = await personAPI.create({
    name: "Jim",
  });
  await personAPI.create({
    name: "Sally",
  });
  const rows = await personAPI.findMany({
    name: {
      equals: "Jim",
    },
  });
  expect(rows.length).toEqual(1);
  expect(rows[0].id).toEqual(p.id);
});

test("ModelAPI.findMany - notEquals", async () => {
  const p = await personAPI.create({
    name: "Jim",
  });
  await personAPI.create({
    name: "Sally",
  });
  const rows = await personAPI.findMany({
    name: {
      notEquals: "Sally",
    },
  });
  expect(rows.length).toEqual(1);
  expect(rows[0].id).toEqual(p.id);
});

test("ModelAPI.findMany - complex query", async () => {
  const p = await personAPI.create({
    name: "Jake",
    favouriteNumber: 8,
    date: new Date("2021-12-31"),
  });
  await personAPI.create({
    name: "Jane",
    favouriteNumber: 12,
    date: new Date("2022-01-11"),
  });
  const p2 = await personAPI.create({
    name: "Billy",
    favouriteNumber: 16,
    date: new Date("2022-01-05"),
  });

  const rows = await personAPI
    // Will match Jake
    .where({
      name: {
        startsWith: "J",
        endsWith: "e",
      },
      favouriteNumber: {
        lessThan: 10,
      },
    })
    // Will match Billy
    .orWhere({
      date: {
        after: new Date("2022-01-01"),
        before: new Date("2022-01-10"),
      },
    })
    .findMany();
  expect(rows.length).toEqual(2);
  expect(rows.map((x) => x.id).sort()).toEqual([p.id, p2.id].sort());
});

test("ModelAPI.findMany - relationships - one to many", async () => {
  const person = await personAPI.create({
    name: "Jim",
  });
  const person2 = await personAPI.create({
    name: "Bob",
  });
  const post1 = await postAPI.create({
    title: "My First Post",
    authorId: person.id,
  });
  const post2 = await postAPI.create({
    title: "My Second Post",
    authorId: person.id,
  });
  await postAPI.create({
    title: "My Third Post",
    authorId: person2.id,
  });

  const posts = await postAPI.findMany({
    author: {
      name: "Jim",
    },
  });
  expect(posts.length).toEqual(2);
  expect(posts.map((x) => x.id).sort()).toEqual([post1.id, post2.id].sort());
});

test("ModelAPI.findMany - relationships - many to one", async () => {
  const person = await personAPI.create({
    name: "Jim",
  });
  await postAPI.create({
    title: "My First Post",
    authorId: person.id,
  });
  await postAPI.create({
    title: "My Second Post",
    authorId: person.id,
  });
  await postAPI.create({
    title: "My Second Post",
    authorId: person.id,
  });

  const people = await personAPI.findMany({
    posts: {
      title: {
        startsWith: "My ",
        endsWith: " Post",
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
    name: "Jim",
  });
  const person2 = await personAPI.create({
    name: "Bob",
  });
  const post1 = await postAPI.create({
    title: "My First Post",
    authorId: person.id,
  });
  const post2 = await postAPI.create({
    title: "My Second Post",
    authorId: person2.id,
  });

  const posts = await postAPI
    .where({
      author: {
        name: "Jim",
      },
    })
    .orWhere({
      author: {
        name: "Bob",
      },
    })
    .findMany();

  expect(posts.length).toEqual(2);
  expect(posts.map((x) => x.id).sort()).toEqual([post1.id, post2.id].sort());
});

test("ModelAPI.update", async () => {
  let jim = await personAPI.create({
    name: "Jim",
    married: false,
    favouriteNumber: 10,
  });
  let bob = await personAPI.create({
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
    'null value in column "name" violates not-null constraint'
  );
});

test("ModelAPI.delete", async () => {
  const jim = await personAPI.create({
    name: "Jim",
  });
  const id = jim.id;
  const deletedId = await personAPI.delete({
    name: "Jim",
  });

  expect(deletedId).toEqual(id);
  await expect(personAPI.findOne({ id })).resolves.toEqual(null);
});
