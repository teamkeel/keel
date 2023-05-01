import { format, validate, completions, getDefinition } from "./index";
import { test, expect } from "vitest";

const configFile = `
environment:
  default:
    - name: "TEST"
      value: "test"

  staging:
    - name: "TEST_2"
      value: "test2"

secrets:
  - name: API_KEY
    required:
      - "production"
`;

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
  const result = await completions(
    schema,
    {
      line: 3,
      column: 16,
    },
    configFile
  );

  expect(result.completions.map((x) => x.label)).toContain("Text");
});

test("validate", async () => {
  const schema = `model Person {
    fields { 
        name Foo
    }
}`;
  const { errors } = await validate(schema, configFile);

  expect(errors[0].message).toEqual("field name has an unsupported type Foo");
});

test("validate - invalid schema", async () => {
  const schema = `model Person {
    fields {`;
  const { errors } = await validate(schema, configFile);

  expect(errors[0].code).toEqual("E025");
  expect(errors[0].message).toEqual(` unexpected token "<EOF>" (expected "}")`);
});

test("getDefinition", async () => {
  const result = await getDefinition({
    position: {
      line: 7,
      column: 21,
      offset: 0,
      filename: "myschema.keel",
    },
    schemaFiles: [
      {
        filename: "myschema.keel",
        contents: `
model Person {
  fields {
    name Text
  }
  operations {
    list getPeople(name) 
  }
}
        `,
      },
    ],
  });

  expect(result).toEqual({
    function: null,
    schema: {
      filename: "myschema.keel",
      line: 4,
      column: 5,
    },
  });
});

test("getDefinition - no result", async () => {
  const result = await getDefinition({
    position: {
      line: 1,
      column: 1,
      offset: 0,
      filename: "myschema.keel",
    },
    schemaFiles: [
      {
        filename: "myschema.keel",
        contents: `
model Person {
  fields {
    name Text
  }
}
        `,
      },
    ],
  });

  expect(result).toBeNull();
});
