import { test, expect, actions, Thing } from "@teamkeel/testing";

/* 
  Text Type 
*/

test("text set attribute on optional field - set to goodbye - is goodbye", async () => {
  const { object: thing } = await actions.createTextOnOptional({});
  expect(thing.optionalText).toEqual("goodbye");
});

test("text set attribute on optional field - set to null - is null", async () => {
  const { object: thing } = await actions.createNullTextOnOptional({});
  expect(thing.optionalText).toEqual(null);
});

test("text set attribute on required field - set to goodbye - is goodbye", async () => {
  const { object: thing } = await actions.createTextOnRequired({});
  expect(thing.requiredText).toEqual("goodbye");
});

test("text set attribute from explicit input - set to goodbye - is goodbye", async () => {
  const { object: thing } = await actions.createTextFromExplicitInput({
    explText: "goodbye",
  });
  expect(thing.requiredText).toEqual("goodbye");
});

test("text set attribute from implicit input - set to goodbye - is goodbye", async () => {
  const { object: thing } = await actions.createTextFromImplicitInput({
    requiredText: "goodbye",
  });
  expect(thing.optionalText).toEqual("goodbye");
  expect(thing.requiredText).toEqual("goodbye");
});

/* 
  Number Type 
*/

test("number set attribute on optional field - set to 5 - is 5", async () => {
  const { object: thing } = await actions.createNumberOnOptional({});
  expect(thing.optionalNumber).toEqual(5);
});

test("number set attribute on optional field - set to 1 - is null", async () => {
  const { object: thing } = await actions.createNullNumberOnOptional({});
  expect(thing.optionalNumber).toEqual(null);
});

test("number set attribute on required field - set to 5 - is 5", async () => {
  const { object: thing } = await actions.createNumberOnRequired({});
  expect(thing.requiredNumber).toEqual(5);
});

test("number set attribute from explicit input - set to 5 - is 5", async () => {
  const { object: thing } = await actions.createNumberFromExplicitInput({
    explNumber: 5,
  });
  expect(thing.requiredNumber).toEqual(5);
});

test("number set attribute from implicit input - set to 5 - is 5", async () => {
  const { object: thing } = await actions.createNumberFromImplicitInput({
    requiredNumber: 5,
  });
  expect(thing.optionalNumber).toEqual(5);
  expect(thing.requiredNumber).toEqual(5);
});

/* 
  Boolean Type 
*/

test("boolean set attribute on optional field - set to true - is true", async () => {
  const { object: thing } = await actions.createBooleanOnOptional({});
  expect(thing.optionalBoolean).toEqual(true);
});

test("boolean set attribute on optional field - set to null - is null", async () => {
  const { object: thing } = await actions.createNullBooleanOnOptional({});
  expect(thing.optionalBoolean).toEqual(null);
});

test("boolean set attribute on required field - set to true - is true", async () => {
  const { object: thing } = await actions.createBooleanOnRequired({});
  expect(thing.requiredBoolean).toEqual(true);
});

test("boolean set attribute from explicit input - set to true - is true", async () => {
  const { object: thing } = await actions.createBooleanFromExplicitInput({
    explBoolean: true,
  });
  expect(thing.requiredBoolean).toEqual(true);
});

test("boolean set attribute from implicit input - set to true - is true", async () => {
  const { object: thing } = await actions.createBooleanFromImplicitInput({
    requiredBoolean: true,
  });
  expect(thing.optionalBoolean).toEqual(true);
  expect(thing.requiredBoolean).toEqual(true);
});

/* 
  Enum Type 
  Use enum type: https://linear.app/keel/issue/DEV-204/export-enum-types-from-teamkeeltesting
*/

test("enum set attribute on optional field - set to TypeTwo - is TypeTwo", async () => {
  const { object: thing } = await actions.createEnumOnOptional({});
  expect(thing.optionalEnum).toEqual("TypeTwo");
});

test("enum set attribute on optional field - set to null - is null", async () => {
  const { object: thing } = await actions.createNullEnumOnOptional({});
  expect(thing.optionalEnum).toEqual(null);
});

test("enum set attribute on required field - set to TypeTwo - is TypeTwo", async () => {
  const { object: thing } = await actions.createEnumOnRequired({});
  expect(thing.requiredEnum).toEqual("TypeTwo");
});

test("enum set attribute from explicit input - set to TypeTwo - is TypeTwo", async () => {
  const { object: thing } = await actions.createEnumFromExplicitInput({
    explEnum: "TypeTwo",
  });
  expect(thing.requiredEnum).toEqual("TypeTwo");
});

test("enum set attribute from implicit input - set to TypeTwo - is TypeTwo", async () => {
  const { object: thing } = await actions.createEnumFromImplicitInput({
    requiredEnum: "TypeTwo",
  });
  expect(thing.optionalEnum).toEqual("TypeTwo");
  expect(thing.requiredEnum).toEqual("TypeTwo");
});
