import { format, validate, completions } from "./index";
import { test, expect } from "vitest";

test("format", async () => {
  const schema = `model Person { fields { name Text } }`;
  const formatted = await format(schema);
  expect(formatted).toEqual(`model Person {
    fields {
        name Text
    }
}
`);
});

test("format - invalid schema", async () => {
  const schema = `model    Person    {`;
  const formatted = await format(schema);
  expect(formatted).toEqual(schema);
});

test("completions", async () => {
  const schema = `model Person {
    fields { 
        name Te 
    }
}`;
  const result = await completions(schema, {
    line: 3,
    column: 16,
  });

  expect(result.completions.map((x) => x.label)).toContain("Text");
});

test("validate", async () => {
  const schema = `model Person {
    fields { 
        name Foo
    }
}`;
  const { errors } = await validate(schema);

  expect(errors[0].message).toEqual("field name has an unsupported type Foo");
});

test("validate - invalid schema", async () => {
  const schema = `model Person {
    fields {`;
  const { errors } = await validate(schema);

  expect(errors[0].code).toEqual("E025");
  expect(errors[0].message).toEqual(` unexpected token "<EOF>" (expected "}")`);
});
