import { z } from "zod";
import { Conditions, OrderClauses } from "types";
import {
  buildSelectStatement,
  buildDeleteStatement,
  buildCreateStatement,
  buildUpdateStatement,
} from "./";

interface Test {
  foo: string;
  bar: number;
}

test("buildSelectStatement", () => {
  const zod = z.object({
    foo: z.string(),
    bar: z.number(),
  });
  const query = buildSelectStatement<Test>("test", zod, [
    {
      foo: {
        startsWith: "bar",
      },
    },
    {
      bar: {
        greaterThan: 1,
      },
    },
  ] as Conditions<Test>[]);

  const { sql, values } = query;

  expect(sql).toEqual(
    'SELECT * FROM "test" WHERE ("test"."foo" ILIKE $1) OR ("test"."bar" > $2)'
  );

  expect(values).toEqual(["bar%", 1]);
});

test("buildDeleteStatement", () => {
  const id = "jdssjdjsjj";
  const zod = z.object({
    foo: z.string(),
    bar: z.number(),
  });
  const { sql, values } = buildDeleteStatement<Test>("test", zod, id);

  expect(sql).toEqual('DELETE FROM "test" WHERE id = $1 RETURNING id');

  expect(values).toEqual([id]);
});

test("buildCreateStatement", () => {
  const t: Test = {
    foo: "bar",
    bar: 1,
  };
  const zod = z.object({
    foo: z.string(),
    bar: z.number(),
  });
  const { sql, values } = buildCreateStatement<Test>("test", zod, t);

  expect(sql).toEqual(`
    INSERT INTO "test" ("foo", "bar")
    VALUES ($1, $2)
    RETURNING id`);

  expect(values).toEqual(["bar", 1]);
});

test("buildUpdateStatement", () => {
  const id = "18jsjsj";
  const t: Test = {
    foo: "bar",
    bar: 1,
  };
  const zod = z.object({
    foo: z.string(),
    bar: z.number(),
  });
  const { sql, values } = buildUpdateStatement<Test>("test", zod, id, t);

  expect(sql).toEqual('UPDATE "test" SET "foo" = $1,"bar" = $2 WHERE id = $3');

  expect(values).toEqual(["bar", 1, id]);
});

test("testLimit", () => {
  const zod = z.object({
    foo: z.string(),
    bar: z.number(),
  });
  const query = buildSelectStatement<Test>(
    "test",
    zod,
    [
      {
        foo: {
          startsWith: "bar",
        },
      },
    ] as Conditions<Test>[],
    undefined,
    1
  );

  const { sql, values } = query;

  expect(sql).toEqual(
    'SELECT * FROM "test" WHERE ("test"."foo" ILIKE $1) LIMIT $2'
  );

  expect(values).toEqual(["bar%", 1]);
});

test("testOrder", () => {
  const zod = z.object({
    foo: z.string(),
    bar: z.number(),
  });
  const query = buildSelectStatement<Test>(
    "test",
    zod,
    [
      {
        foo: {
          startsWith: "bar",
        },
      },
    ] as Conditions<Test>[],
    {
      foo: "ASC",
    } as OrderClauses<Test>
  );

  const { sql, values } = query;

  expect(sql).toEqual(
    'SELECT * FROM "test" WHERE ("test"."foo" ILIKE $1) ORDER BY $2'
  );

  expect(values).toEqual(["bar%", "foo ASC"]);
});
