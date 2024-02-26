import { actions, models, resetDatabase } from "@teamkeel/testing";
import { MyEnum } from "@teamkeel/sdk";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

test("array expressions - text", async () => {
  const thing1 = await actions.createThing({
    array1: ["Keel", "Weave"],
    array2: ["Keel", "Weave"],
    value: "Keel",
  });

  const thing2 = await actions.createThing({
    array1: ["Weave", "Keel"],
    array2: ["Keel", "Weave"],
    value: "Not Keel",
  });

  const thing3 = await actions.createThing({
    array1: ["Weave", "Keel"],
    array2: ["Weave", "Keel"],
    value: "Keel",
  });

  const thing4 = await actions.createThing({
    array1: null,
    array2: ["Weave", "Keel"],
  });

  const thing5 = await actions.createThing({
    array1: ["Weave", "Keel"],
    array2: null,
  });

  const thing6 = await actions.createThing({
    array1: null,
    array2: null,
  });

  const thing7 = await actions.createThing({
    array1: [],
    array2: ["Keel", "Weave"],
  });

  const thing8 = await actions.createThing({
    array1: ["Keel", "Weave"],
    array2: [],
  });

  const thing9 = await actions.createThing({
    array1: [],
    array2: [],
  });

  const thing10 = await actions.createThing({
    array1: ["something", "else"],
    array2: ["else", "something"],
  });

  const things = await actions.listEqualToLiteral();
  expect(things.results).toHaveLength(2);
  expect(things.results[0].id).toEqual(thing1.id);
  expect(things.results[1].id).toEqual(thing8.id);

  const things2 = await actions.listEqualToField();
  expect(things2.results).toHaveLength(4);
  expect(things2.results[0].id).toEqual(thing1.id);
  expect(things2.results[1].id).toEqual(thing3.id);
  expect(things2.results[2].id).toEqual(thing6.id);
  expect(things2.results[3].id).toEqual(thing9.id);

  const things3 = await actions.listEqualToNull();
  expect(things3.results).toHaveLength(2);
  expect(things3.results[0].id).toEqual(thing4.id);
  expect(things3.results[1].id).toEqual(thing6.id);

  const things4 = await actions.listLiteralInArrayField();
  expect(things4.results).toHaveLength(5);
  expect(things4.results[0].id).toEqual(thing1.id);
  expect(things4.results[1].id).toEqual(thing2.id);
  expect(things4.results[2].id).toEqual(thing3.id);
  expect(things4.results[3].id).toEqual(thing5.id);
  expect(things4.results[4].id).toEqual(thing8.id);

  const things5 = await actions.listLiteralNotInArrayField();
  expect(things5.results).toHaveLength(5);
  expect(things5.results[0].id).toEqual(thing4.id);
  expect(things5.results[1].id).toEqual(thing6.id);
  expect(things5.results[2].id).toEqual(thing7.id);
  expect(things5.results[3].id).toEqual(thing9.id);
  expect(things5.results[4].id).toEqual(thing10.id);

  const things6 = await actions.listFieldInArrayField();
  expect(things6.results).toHaveLength(2);
  expect(things6.results[0].id).toEqual(thing1.id);
  expect(things6.results[1].id).toEqual(thing3.id);
});

test("array expressions - enums", async () => {
  const thing1 = await actions.createEnumThing({
    array1: [MyEnum.One, MyEnum.Two],
    array2: [MyEnum.One, MyEnum.Two],
    value: MyEnum.One,
  });

  const thing2 = await actions.createEnumThing({
    array1: [MyEnum.Two, MyEnum.One],
    array2: [MyEnum.One, MyEnum.Two],
    value: MyEnum.Three,
  });

  const thing3 = await actions.createEnumThing({
    array1: [MyEnum.Two, MyEnum.One],
    array2: [MyEnum.Two, MyEnum.One],
    value: MyEnum.One,
  });

  const thing4 = await actions.createEnumThing({
    array1: null,
    array2: [MyEnum.Two, MyEnum.One],
  });

  const thing5 = await actions.createEnumThing({
    array1: [MyEnum.Two, MyEnum.One],
    array2: null,
  });

  const thing6 = await actions.createEnumThing({
    array1: null,
    array2: null,
  });

  const thing7 = await actions.createEnumThing({
    array1: [],
    array2: [MyEnum.One, MyEnum.Two],
  });

  const thing8 = await actions.createEnumThing({
    array1: [MyEnum.One, MyEnum.Two],
    array2: [],
  });

  const thing9 = await actions.createEnumThing({
    array1: [],
    array2: [],
  });

  const thing10 = await actions.createEnumThing({
    array1: [MyEnum.Three, MyEnum.Four],
    array2: [MyEnum.Four, MyEnum.Three],
  });

  const things = await actions.listEnumEqualToLiteral();
  expect(things.results).toHaveLength(2);
  expect(things.results[0].id).toEqual(thing1.id);
  expect(things.results[1].id).toEqual(thing8.id);

  const things2 = await actions.listEnumEqualToField();
  expect(things2.results).toHaveLength(4);
  expect(things2.results[0].id).toEqual(thing1.id);
  expect(things2.results[1].id).toEqual(thing3.id);
  expect(things2.results[2].id).toEqual(thing6.id);
  expect(things2.results[3].id).toEqual(thing9.id);

  const things3 = await actions.listEnumEqualToNull();
  expect(things3.results).toHaveLength(2);
  expect(things3.results[0].id).toEqual(thing4.id);
  expect(things3.results[1].id).toEqual(thing6.id);

  const things4 = await actions.listEnumLiteralInArrayField();
  expect(things4.results).toHaveLength(5);
  expect(things4.results[0].id).toEqual(thing1.id);
  expect(things4.results[1].id).toEqual(thing2.id);
  expect(things4.results[2].id).toEqual(thing3.id);
  expect(things4.results[3].id).toEqual(thing5.id);
  expect(things4.results[4].id).toEqual(thing8.id);

  const things5 = await actions.listEnumLiteralNotInArrayField();
  expect(things5.results).toHaveLength(5);
  expect(things5.results[0].id).toEqual(thing4.id);
  expect(things5.results[1].id).toEqual(thing6.id);
  expect(things5.results[2].id).toEqual(thing7.id);
  expect(things5.results[3].id).toEqual(thing9.id);
  expect(things5.results[4].id).toEqual(thing10.id);

  const things6 = await actions.listEnumFieldInArrayField();
  expect(things6.results).toHaveLength(2);
  expect(things6.results[0].id).toEqual(thing1.id);
  expect(things6.results[1].id).toEqual(thing3.id);
});
