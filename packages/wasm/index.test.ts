import { keel } from "./index";
import { test, expect } from "vitest";

test("format", async () => {
  const api = keel();
  const schema = `model Person { fields { name Text } }`;
  const formatted = await api.format(schema);
  expect(formatted).toEqual(`model Person {
    fields {
        name Text
    }
}
`);
});

test("completions", async () => {
  const api = keel();
  const schema = `model Person {
    fields { 
        name Te 
    }
}`;
  const { completions } = await api.completions(schema, {
    line: 3,
    column: 16,
  });

  expect(completions.map((x) => x.label)).toContain("Text");
});

test("validate", async () => {
  const api = keel();
  const schema = `model Person {
    fields { 
        name Foo
    }
}`;
  const { errors } = await api.validate(schema);

  expect(errors[0].message).toEqual("field name has an unsupported type Foo");
});
