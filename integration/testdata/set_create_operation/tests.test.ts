import { actions, resetDatabase, models } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";
import { ThingType } from "@teamkeel/sdk";

beforeEach(resetDatabase);

/* 
  Text Type 
*/

test("text set attribute on optional field - set to goodbye - is goodbye", async () => {
  const thing = await actions.createTextOnOptional({});
  expect(thing.optionalText).toEqual("goodbye");
});

test("text set attribute on optional field - set to null - is null", async () => {
  const thing = await actions.createNullTextOnOptional({});
  expect(thing.optionalText).toEqual(null);
});

test("text set attribute on required field - set to goodbye - is goodbye", async () => {
  const thing = await actions.createTextOnRequired({});
  expect(thing.requiredText).toEqual("goodbye");
});

test("text set attribute from explicit input - set to goodbye - is goodbye", async () => {
  const thing = await actions.createTextFromExplicitInput({
    explText: "goodbye",
  });
  expect(thing.requiredText).toEqual("goodbye");
});

test("text set attribute from implicit input - set to goodbye - is goodbye", async () => {
  const thing = await actions.createTextFromImplicitInput({
    requiredText: "goodbye",
  });
  expect(thing.optionalText).toEqual("goodbye");
  expect(thing.requiredText).toEqual("goodbye");
});

/* 
  Number Type 
*/

test("number set attribute on optional field - set to 5 - is 5", async () => {
  const thing = await actions.createNumberOnOptional({});
  expect(thing.optionalNumber).toEqual(5);
});

test("number set attribute on optional field - set to 1 - is null", async () => {
  const thing = await actions.createNullNumberOnOptional({});
  expect(thing.optionalNumber).toEqual(null);
});

test("number set attribute on required field - set to 5 - is 5", async () => {
  const thing = await actions.createNumberOnRequired({});
  expect(thing.requiredNumber).toEqual(5);
});

test("number set attribute from explicit input - set to 5 - is 5", async () => {
  const thing = await actions.createNumberFromExplicitInput({
    explNumber: 5,
  });
  expect(thing.requiredNumber).toEqual(5);
});

test("number set attribute from implicit input - set to 5 - is 5", async () => {
  const thing = await actions.createNumberFromImplicitInput({
    requiredNumber: 5,
  });
  expect(thing.optionalNumber).toEqual(5);
  expect(thing.requiredNumber).toEqual(5);
});

/* 
  Decimal Type 
*/

test("decimal set attribute on optional field - set to 1.5 - is 1.5", async () => {
  const thing = await actions.createDecimalOnOptional({});
  expect(thing.optionalDecimal).toEqual(1.5);
});

test("decimal set attribute on optional field - set to 1.5 - is null", async () => {
  const thing = await actions.createNullDecimalOnOptional({});
  expect(thing.optionalDecimal).toEqual(null);
});

test("decimal set attribute on required field - set to 1.5 - is 1.5", async () => {
  const thing = await actions.createDecimalOnRequired({});
  expect(thing.requiredDecimal).toEqual(1.5);
});

test("decimal set attribute from explicit input - set to 2.5 - is 2.5", async () => {
  const thing = await actions.createDecimalFromExplicitInput({
    explDecimal: 2.5,
  });
  expect(thing.requiredDecimal).toEqual(2.5);
});

test("decimal set attribute from implicit input - set to 2.5 - is 2.5", async () => {
  const thing = await actions.createDecimalFromImplicitInput({
    requiredDecimal: 2.5,
  });
  expect(thing.optionalDecimal).toEqual(2.5);
  expect(thing.requiredDecimal).toEqual(2.5);
});

/* 
  Boolean Type 
*/

test("boolean set attribute on optional field - set to true - is true", async () => {
  const thing = await actions.createBooleanOnOptional({});
  expect(thing.optionalBoolean).toEqual(true);
});

test("boolean set attribute on optional field - set to null - is null", async () => {
  const thing = await actions.createNullBooleanOnOptional({});
  expect(thing.optionalBoolean).toEqual(null);
});

test("boolean set attribute on required field - set to true - is true", async () => {
  const thing = await actions.createBooleanOnRequired({});
  expect(thing.requiredBoolean).toEqual(true);
});

test("boolean set attribute from explicit input - set to true - is true", async () => {
  const thing = await actions.createBooleanFromExplicitInput({
    explBoolean: true,
  });
  expect(thing.requiredBoolean).toEqual(true);
});

test("boolean set attribute from implicit input - set to true - is true", async () => {
  const thing = await actions.createBooleanFromImplicitInput({
    requiredBoolean: true,
  });
  expect(thing.optionalBoolean).toEqual(true);
  expect(thing.requiredBoolean).toEqual(true);
});

/* 
  Duration Type 
*/
test("duration set attribute on optional field - set to P1D - is P1D", async () => {
  const thing = await actions.createDurationOnOptional({});
  expect(thing.optionalDuration).toEqual("P1D");
});

test("duration set attribute on optional field - set to null - is null", async () => {
  const thing = await actions.createNullDurationOnOptional({});
  expect(thing.optionalDuration).toEqual(null);
});

test("duration set attribute on required field - set to P1D - is P1D", async () => {
  const thing = await actions.createDurationOnRequired({});
  expect(thing.requiredDuration).toEqual("P1D");
});

test("duration set attribute from explicit input - set to P2D - is P2D", async () => {
  const thing = await actions.createDurationFromExplicitInput({
    explDuration: "P2D",
  });
  expect(thing.requiredDuration).toEqual("P2D");
});

test("duration set attribute from implicit input - set to P2D - is P2D", async () => {
  const thing = await actions.createDurationFromImplicitInput({
    requiredDuration: "P2D",
  });
  expect(thing.optionalDuration).toEqual("P2D");
  expect(thing.requiredDuration).toEqual("P2D");
});

/* 
  Enum Type 
  Use enum type: https://linear.app/keel/issue/DEV-204/export-enum-types-from-teamkeeltesting
*/

test("enum set attribute on optional field - set to TypeTwo - is TypeTwo", async () => {
  const thing = await actions.createEnumOnOptional({});
  expect(thing.optionalEnum).toEqual(ThingType.TypeTwo);
});

test("enum set attribute on optional field - set to null - is null", async () => {
  const thing = await actions.createNullEnumOnOptional({});
  expect(thing.optionalEnum).toEqual(null);
});

test("enum set attribute on required field - set to TypeTwo - is TypeTwo", async () => {
  const thing = await actions.createEnumOnRequired({});
  expect(thing.requiredEnum).toEqual(ThingType.TypeTwo);
});

test("enum set attribute from explicit input - set to TypeTwo - is TypeTwo", async () => {
  const thing = await actions.createEnumFromExplicitInput({
    explEnum: ThingType.TypeTwo,
  });
  expect(thing.requiredEnum).toEqual(ThingType.TypeTwo);
});

test("enum set attribute from implicit input - set to TypeTwo - is TypeTwo", async () => {
  const thing = await actions.createEnumFromImplicitInput({
    requiredEnum: ThingType.TypeTwo,
  });
  expect(thing.optionalEnum).toEqual(ThingType.TypeTwo);
  expect(thing.requiredEnum).toEqual(ThingType.TypeTwo);
});

/*
  Model Type
*/

test("model set attribute from explicit input - set to parent - has parent", async () => {
  const parent = await models.parent.create({
    name: "Keelson",
  });

  const thing = await actions.createParentFromExplicitInput({
    explParent: parent.id,
  });
  expect(thing.optionalParentId).toEqual(parent.id);
});

test("model set attribute on optional foreign key ID field - set to null - is null", async () => {
  const thing = await actions.createNullParentId({});
  expect(thing.optionalParentId).toEqual(null);
});

test("model set attribute on optional field - set to null - is null", async () => {
  const thing = await actions.createNullParent({});
  expect(thing.optionalParentId).toEqual(null);
});
