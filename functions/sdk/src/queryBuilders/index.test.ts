import { Conditions, OrderClauses } from "types";
import {
  buildSelectStatement,
  buildDeleteStatement,
  buildCreateStatement,
  buildUpdateStatement,
  transformValue
} from "./";

interface Test {
  foo: string;
  bar: number;
}

test("buildSelectStatement", () => {
  const query = buildSelectStatement<Test>("test", [
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

  const { sql, values } = buildDeleteStatement<Test>("test", id);

  expect(sql).toEqual('DELETE FROM "test" WHERE id = $1 RETURNING id');

  expect(values).toEqual([id]);
});

test("buildCreateStatement", () => {
  const t: Test = {
    foo: "bar",
    bar: 1,
  };

  const { sql, values } = buildCreateStatement<Test>("test", t);

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

  const { sql, values } = buildUpdateStatement<Test>("test", id, t);

  expect(sql).toEqual('UPDATE "test" SET "foo" = $1,"bar" = $2 WHERE id = $3');

  expect(values).toEqual(["bar", 1, id]);
});

test("testLimit", () => {
  const query = buildSelectStatement<Test>(
    "test",
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
  const query = buildSelectStatement<Test>(
    "test",
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

describe('transformValue', () => {
  it('converts Date objects to ISO8601', () => {
    const d = new Date(2020, 3, 1);

    const result = transformValue(d);

    expect(result).toEqual("2020-03-31T23:00:00.000Z")
  });
  
  it('returns the original value for everything else', () => {
    const primitives = [1, 's', true, undefined, null];

    primitives.forEach((p) => {
      expect(transformValue(p)).toEqual(p);
    });
  });
});
