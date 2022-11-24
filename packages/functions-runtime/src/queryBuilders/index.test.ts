import { Conditions, OrderClauses } from "types";
import {
  buildSelectStatement,
  buildDeleteStatement,
  buildCreateStatement,
  buildUpdateStatement,
} from "./";
import { SqlQueryParts } from "../db/query";

interface Test {
  foo: string;
  bar: number;
}

function toPreparedStatement(query: SqlQueryParts): {
  sql: string;
  values: any[];
} {
  let nextInterpolationIndex = 1;
  let values = [];
  const sql = query
    .map((queryPart) => {
      switch (queryPart.type) {
        case "sql":
          return queryPart.value;
        case "input":
          values.push(queryPart.value);
          return `$${nextInterpolationIndex++}`;
      }
    })
    .join(" ");
  return { sql, values };
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

  const { sql, values } = toPreparedStatement(query);

  expect(sql).toEqual(
    'SELECT * FROM "test" WHERE ( "test"."foo" ILIKE $1 ) OR ( "test"."bar" > $2 )'
  );

  expect(values).toEqual(["bar%", 1]);
});

test("buildDeleteStatement", () => {
  const id = "jdssjdjsjj";

  const { sql, values } = toPreparedStatement(
    buildDeleteStatement<Test>("test", id)
  );

  expect(sql).toEqual('DELETE FROM "test" WHERE id = $1 RETURNING id');

  expect(values).toEqual([id]);
});

test("buildCreateStatement", () => {
  const t: Test = {
    foo: "bar",
    bar: 1,
  };

  const { sql, values } = toPreparedStatement(
    buildCreateStatement<Test>("test", t)
  );

  expect(sql).toEqual(
    `INSERT INTO "test" ( "foo" , "bar" ) VALUES ( $1 , $2 ) RETURNING id`
  );

  expect(values).toEqual(["bar", 1]);
});

test("buildUpdateStatement", () => {
  const id = "18jsjsj";
  const t: Test = {
    foo: "bar",
    bar: 1,
  };

  const { sql, values } = toPreparedStatement(
    buildUpdateStatement<Test>("test", id, t)
  );

  expect(sql).toEqual(
    'UPDATE "test" SET "foo" = $1 , "bar" = $2 WHERE id = $3'
  );

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

  const { sql, values } = toPreparedStatement(query);

  expect(sql).toEqual(
    'SELECT * FROM "test" WHERE ( "test"."foo" ILIKE $1 ) LIMIT $2'
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

  const { sql, values } = toPreparedStatement(query);

  expect(sql).toEqual(
    'SELECT * FROM "test" WHERE ( "test"."foo" ILIKE $1 ) ORDER BY foo ASC'
  );

  expect(values).toEqual(["bar%"]);
});
