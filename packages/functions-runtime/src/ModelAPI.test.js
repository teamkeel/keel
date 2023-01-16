import { test, expect, beforeEach } from "vitest";
const { ModelAPI } = require("./ModelAPI");
const { sql } = require("kysely");
const { getDatabase } = require("./database");
const KSUID = require("ksuid");

process.env.DB_CONN_TYPE = "pg";
process.env.DB_CONN = `postgresql://postgres:postgres@localhost:5432/functions-runtime`;

let api;

beforeEach(async () => {
  const db = getDatabase();

  await sql`
  DROP TABLE IF EXISTS model_api_test;
  CREATE TABLE model_api_test(
      id               text PRIMARY KEY,
      name             text UNIQUE,
      married          boolean,
      favourite_number integer,
      date             timestamp
  );
  `.execute(db);

  api = new ModelAPI(
    "model_api_test",
    () => {
      return {
        id: KSUID.randomSync().string,
        date: new Date("2022-01-01"),
      };
    },
    db
  );
});

test("ModelAPI.create", async () => {
  const row = await api.create({
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

test("ModelAPI.create - throws if database constraint fails", async () => {
  const row = await api.create({
    name: "Jim",
    married: false,
    favouriteNumber: 10,
  });
  const promise = api.create({
    id: row.id,
    name: "Jim",
    married: false,
    favouriteNumber: 10,
  });
  await expect(promise).rejects.toThrow(
    `duplicate key value violates unique constraint "model_api_test_pkey"`
  );
});

test("ModelAPI.findOne", async () => {
  const created = await api.create({
    name: "Jim",
    married: false,
    favouriteNumber: 10,
  });
  const row = await api.findOne({
    id: created.id,
  });
  expect(row).toEqual(created);
});

test("ModelAPI.findOne - return null if not found", async () => {
  const row = await api.findOne({
    id: "doesntexist",
  });
  expect(row).toEqual(null);
});

test("ModelAPI.findMany", async () => {
  const jim = await api.create({
    name: "Jim",
    married: false,
    favouriteNumber: 10,
  });
  const bob = await api.create({
    name: "Bob",
    married: true,
    favouriteNumber: 11,
  });
  const sally = await api.create({
    name: "Sally",
    married: true,
    favouriteNumber: 12,
  });
  const rows = await api.findMany({
    married: true,
  });
  expect(rows.length).toEqual(2);
  expect(rows.map((x) => x.id).sort()).toEqual([bob.id, sally.id].sort());
});

test("ModelAPI.findMany - startsWith", async () => {
  const jim = await api.create({
    name: "Jim",
  });
  await api.create({
    name: "Bob",
  });
  const rows = await api.findMany({
    name: {
      startsWith: "Ji",
    },
  });
  expect(rows.length).toEqual(1);
  expect(rows[0].id).toEqual(jim.id);
});

test("ModelAPI.findMany - endsWith", async () => {
  const jim = await api.create({
    name: "Jim",
  });
  await api.create({
    name: "Bob",
  });
  const rows = await api.findMany({
    name: {
      endsWith: "im",
    },
  });
  expect(rows.length).toEqual(1);
  expect(rows[0].id).toEqual(jim.id);
});

test("ModelAPI.findMany - contains", async () => {
  const billy = await api.create({
    name: "Billy",
  });
  const sally = await api.create({
    name: "Sally",
  });
  await api.create({
    name: "Jim",
  });
  const rows = await api.findMany({
    name: {
      contains: "ll",
    },
  });
  expect(rows.length).toEqual(2);
  expect(rows.map((x) => x.id).sort()).toEqual([billy.id, sally.id].sort());
});

test("ModelAPI.findMany - oneOf", async () => {
  const billy = await api.create({
    name: "Billy",
  });
  const sally = await api.create({
    name: "Sally",
  });
  await api.create({
    name: "Jim",
  });
  const rows = await api.findMany({
    name: {
      oneOf: ["Billy", "Sally"],
    },
  });
  expect(rows.length).toEqual(2);
  expect(rows.map((x) => x.id).sort()).toEqual([billy.id, sally.id].sort());
});

test("ModelAPI.findMany - greaterThan", async () => {
  await api.create({
    favouriteNumber: 1,
  });
  const p = await api.create({
    favouriteNumber: 2,
  });
  const rows = await api.findMany({
    favouriteNumber: {
      greaterThan: 1,
    },
  });
  expect(rows.length).toEqual(1);
  expect(rows[0].id).toEqual(p.id);
});

test("ModelAPI.findMany - greaterThanOrEquals", async () => {
  await api.create({
    favouriteNumber: 1,
  });
  const p = await api.create({
    favouriteNumber: 2,
  });
  const p2 = await api.create({
    favouriteNumber: 3,
  });
  const rows = await api.findMany({
    favouriteNumber: {
      greaterThanOrEquals: 2,
    },
  });
  expect(rows.length).toEqual(2);
  expect(rows.map((x) => x.id).sort()).toEqual([p.id, p2.id].sort());
});

test("ModelAPI.findMany - lessThan", async () => {
  const p = await api.create({
    favouriteNumber: 1,
  });
  await api.create({
    favouriteNumber: 2,
  });
  const rows = await api.findMany({
    favouriteNumber: {
      lessThan: 2,
    },
  });
  expect(rows.length).toEqual(1);
  expect(rows[0].id).toEqual(p.id);
});

test("ModelAPI.findMany - lessThanOrEquals", async () => {
  const p = await api.create({
    favouriteNumber: 1,
  });
  const p2 = await api.create({
    favouriteNumber: 2,
  });
  await api.create({
    favouriteNumber: 3,
  });
  const rows = await api.findMany({
    favouriteNumber: {
      lessThanOrEquals: 2,
    },
  });
  expect(rows.length).toEqual(2);
  expect(rows.map((x) => x.id).sort()).toEqual([p.id, p2.id].sort());
});

test("ModelAPI.findMany - before", async () => {
  const p = await api.create({
    date: new Date("2022-01-01"),
  });
  await api.create({
    date: new Date("2022-01-02"),
  });
  const rows = await api.findMany({
    date: {
      before: new Date("2022-01-02"),
    },
  });
  expect(rows.length).toEqual(1);
  expect(rows[0].id).toEqual(p.id);
});

test("ModelAPI.findMany - onOrBefore", async () => {
  const p = await api.create({
    date: new Date("2022-01-01"),
  });
  const p2 = await api.create({
    date: new Date("2022-01-02"),
  });
  await api.create({
    date: new Date("2022-01-03"),
  });
  const rows = await api.findMany({
    date: {
      onOrBefore: new Date("2022-01-02"),
    },
  });
  expect(rows.length).toEqual(2);
  expect(rows.map((x) => x.id).sort()).toEqual([p.id, p2.id].sort());
});

test("ModelAPI.findMany - after", async () => {
  await api.create({
    date: new Date("2022-01-01"),
  });
  const p = await api.create({
    date: new Date("2022-01-02"),
  });
  const rows = await api.findMany({
    date: {
      after: new Date("2022-01-01"),
    },
  });
  expect(rows.length).toEqual(1);
  expect(rows[0].id).toEqual(p.id);
});

test("ModelAPI.findMany - onOrAfter", async () => {
  await api.create({
    date: new Date("2022-01-01"),
  });
  const p = await api.create({
    date: new Date("2022-01-02"),
  });
  const p2 = await api.create({
    date: new Date("2022-01-03"),
  });
  const rows = await api.findMany({
    date: {
      onOrAfter: new Date("2022-01-02"),
    },
  });
  expect(rows.length).toEqual(2);
  expect(rows.map((x) => x.id).sort()).toEqual([p.id, p2.id].sort());
});

test("ModelAPI.findMany - equals", async () => {
  const p = await api.create({
    name: "Jim",
  });
  await api.create({
    name: "Sally",
  });
  const rows = await api.findMany({
    name: {
      equals: "Jim",
    },
  });
  expect(rows.length).toEqual(1);
  expect(rows[0].id).toEqual(p.id);
});

test("ModelAPI.findMany - notEquals", async () => {
  const p = await api.create({
    name: "Jim",
  });
  await api.create({
    name: "Sally",
  });
  const rows = await api.findMany({
    name: {
      notEquals: "Sally",
    },
  });
  expect(rows.length).toEqual(1);
  expect(rows[0].id).toEqual(p.id);
});

test("ModelAPI.findMany - complex query", async () => {
  const p = await api.create({
    name: "Jake",
    favouriteNumber: 8,
    date: new Date("2021-12-31"),
  });
  await api.create({
    name: "Jane",
    favouriteNumber: 12,
    date: new Date("2022-01-11"),
  });
  const p2 = await api.create({
    name: "Billy",
    favouriteNumber: 16,
    date: new Date("2022-01-05"),
  });

  const rows = await api
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

test("ModelAPI.update", async () => {
  let jim = await api.create({
    name: "Jim",
    married: false,
    favouriteNumber: 10,
  });
  let bob = await api.create({
    name: "Bob",
    married: false,
    favouriteNumber: 11,
  });
  jim = await api.update(
    {
      id: jim.id,
    },
    {
      married: true,
    }
  );
  expect(jim.married).toEqual(true);
  expect(jim.name).toEqual("Jim");

  bob = await api.findOne({ id: bob.id });
  expect(bob.married).toEqual(false);
});

test("ModelAPI.update - throws if not found", async () => {
  const result = api.update(
    {
      id: "doesntexist",
    },
    {
      married: true,
    }
  );
  await expect(result).rejects.toThrow("no result");
});

test("ModelAPI.delete", async () => {
  const jim = await api.create({
    name: "Jim",
  });
  const id = jim.id;
  const deletedId = await api.delete({
    name: "Jim",
  });

  expect(deletedId).toEqual(id);
  await expect(api.findOne({ id })).resolves.toEqual(null);
});
