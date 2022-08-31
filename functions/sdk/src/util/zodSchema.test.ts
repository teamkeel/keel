import { ModelDefinition } from "types";
import { buildZodSchemaFromModelDefinition } from "./zodSchema";

interface Foo {
  bar: string;
  barOne: boolean;
}

test("it does something", () => {
  const def: ModelDefinition<Foo> = {
    bar: "string",
    barOne: "boolean",
  };
  const zodObject = buildZodSchemaFromModelDefinition(def);

  const example = {
    bar: "123",
    barOne: true,
  };

  const goodResult = zodObject.safeParse(example);

  console.log(goodResult);
  expect(goodResult.success).toBe(true);

  const badExample = {
    bar: "title",
    barOne: 123,
  };

  const badResult = zodObject.safeParse(badExample);

  expect(badResult.success).toBe(false);
});
