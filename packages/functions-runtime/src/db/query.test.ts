import { randomUUID } from "crypto";
import {
  rawSql,
  sqlAddSeparator,
  sqlAddSeparatorAndFlatten,
  sqlIdentifier,
  sqlInput,
  sqlInputArray,
} from "./query";
import { test, expect } from "vitest";

test("rawSql", () => {
  const input = randomUUID();
  expect(rawSql(input)).toEqual({
    type: "sql",
    value: input,
  });
});

test("sqlIdentifier", () => {
  const input = randomUUID();
  expect(sqlIdentifier(input)).toEqual({
    type: "sql",
    value: `"${input}"`,
  });

  const input2 = randomUUID();
  expect(sqlIdentifier(input, input2)).toEqual({
    type: "sql",
    value: `"${input}"."${input2}"`,
  });
});

test("sqlInput", () => {
  const input = randomUUID();
  expect(sqlInput(input)).toEqual({
    type: "input",
    value: input,
  });
});

test("sqlInputArray", () => {
  expect(sqlInputArray([])).toEqual([]);
  const input = randomUUID();
  expect(sqlInputArray([input])).toEqual([
    {
      type: "input",
      value: input,
    },
  ]);
  const input2 = randomUUID();
  expect(sqlInputArray([input, input2])).toEqual([
    {
      type: "input",
      value: input,
    },
    {
      type: "input",
      value: input2,
    },
  ]);
});

test("sqlAddSeparator", () => {
  const separators = [rawSql(randomUUID()), sqlInput(randomUUID())];
  for (let sep of separators) {
    expect(sqlAddSeparator([], sep)).toEqual([]);
    const a = sqlInput(randomUUID());
    const b = rawSql(randomUUID());
    const c = sqlInput(randomUUID());
    expect(sqlAddSeparator([a, b, c], sep)).toEqual([a, sep, b, sep, c]);
  }
});

test("sqlAddSeparatorAndFlatten", () => {
  const separators = [rawSql(randomUUID()), sqlInput(randomUUID())];
  for (let sep of separators) {
    expect(sqlAddSeparatorAndFlatten([], sep)).toEqual([]);
    expect(sqlAddSeparatorAndFlatten([[]], sep)).toEqual([]);
    expect(sqlAddSeparatorAndFlatten([[], []], sep)).toEqual([sep]);
    expect(sqlAddSeparatorAndFlatten([[], [], []], sep)).toEqual([sep, sep]);
    const a = sqlInput(randomUUID());
    const b = rawSql(randomUUID());
    const c = sqlInput(randomUUID());
    const d = rawSql(randomUUID());
    expect(sqlAddSeparatorAndFlatten([[a], [b, c, d]], sep)).toEqual([
      a,
      sep,
      b,
      c,
      d,
    ]);
    expect(
      sqlAddSeparatorAndFlatten(
        [
          [a, b],
          [c, d],
        ],
        sep
      )
    ).toEqual([a, b, sep, c, d]);
    expect(sqlAddSeparatorAndFlatten([[a, b], [c], [d]], sep)).toEqual([
      a,
      b,
      sep,
      c,
      sep,
      d,
    ]);
  }
});
