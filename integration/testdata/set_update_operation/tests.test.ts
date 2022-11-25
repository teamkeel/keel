import { test, expect, actions, Thing } from "@teamkeel/testing";

/* 
  Text Type 
*/

test("text set attribute - set to goodbye - is goodbye", async () => {
  const { object: thing } = await actions.create({});
  await actions.updateText({ where: { id: thing.id } });
  const { object: updated } = await actions.get({ id: thing.id });
  expect(updated.text).toEqual("goodbye");
});

test("text set attribute - set to null - is null", async () => {
  const { object: thing } = await actions.create({});
  await actions.updateNullText({ where: { id: thing.id } });
  const { object: updated } = await actions.get({ id: thing.id });
  expect(updated.text).toEqual(null);
});

test("text set attribute from explicit input - set to goodbye - is goodbye", async () => {
  const { object: thing } = await actions.create({});
  const { object: thingUpdated } = await actions.updateTextFromExplicitInput({
    where: { id: thing.id },
    values: { explText: "goodbye" },
  });
  expect(thingUpdated.text).toEqual("goodbye");
});

// https://linear.app/keel/issue/RUN-142/set-with-implicit-inputs-on-update
test("text set attribute from implicit input - set to goodbye - is goodbye", async () => {
  const { object: thing } = await actions.create({});
  const { object: thingUpdated } = await actions.updateTextFromImplicitInput({
    where: { id: thing.id },
    values: { otherText: "goodbye" },
  });
  expect(thingUpdated.text).toEqual("goodbye");
  expect(thingUpdated.otherText).toEqual("goodbye");
});

/* 
  Number Type 
*/

test("number set attribute - set to 5 - is 5", async () => {
  const { object: thing } = await actions.create({});
  await actions.updateNumber({ where: { id: thing.id } });
  const { object: updated } = await actions.get({ id: thing.id });
  expect(updated.number).toEqual(5);
});

test("number set attribute - set to null - is null", async () => {
  const { object: thing } = await actions.create({});
  await actions.updateNullNumber({ where: { id: thing.id } });
  const { object: updated } = await actions.get({ id: thing.id });
  expect(updated.number).toEqual(null);
});

test("number set attribute from explicit input - set to 5 - is 5", async () => {
  const { object: thing } = await actions.create({});
  const { object: thingUpdated } = await actions.updateNumberFromExplicitInput({
    where: { id: thing.id },
    values: { explNumber: 5 },
  });
  expect(thingUpdated.number).toEqual(5);
});

test("number set attribute from implicit input - set to 5 - is 5", async () => {
  const { object: thing } = await actions.create({});
  const { object: thingUpdated } = await actions.updateNumberFromImplicitInput({
    where: { id: thing.id },
    values: { otherNumber: 5 },
  });
  expect(thingUpdated.number).toEqual(5);
  expect(thingUpdated.otherNumber).toEqual(5);
});

/* 
  Boolean Type 
*/

test("boolean set attribute - set to true - is true", async () => {
  const { object: thing } = await actions.create({});
  await actions.updateBoolean({ where: { id: thing.id } });
  const { object: updated } = await actions.get({ id: thing.id });
  expect(updated.boolean).toEqual(true);
});

test("boolean set attribute - set to null - is null", async () => {
  const { object: thing } = await actions.create({});
  await actions.updateNullBoolean({ where: { id: thing.id } });
  const { object: updated } = await actions.get({ id: thing.id });
  expect(updated.boolean).toEqual(null);
});

test("boolean set attribute from explicit input - set to true - is true", async () => {
  const { object: thing } = await actions.create({});
  const { object: thingUpdated } = await actions.updateBooleanFromExplicitInput(
    { where: { id: thing.id }, values: { explBoolean: true } }
  );
  expect(thingUpdated.boolean).toEqual(true);
});

test("boolean set attribute from implicit input - set to true - is true", async () => {
  const { object: thing } = await actions.create({});
  const { object: thingUpdated } = await actions.updateBooleanFromImplicitInput(
    { where: { id: thing.id }, values: { otherBoolean: true } }
  );
  expect(thingUpdated.boolean).toEqual(true);
  expect(thingUpdated.otherBoolean).toEqual(true);
});

/* 
  Enum Type 
*/

test("enum set attribute - set to TypeTwo - is TypeTwo", async () => {
  const { object: thing } = await actions.create({});
  await actions.updateEnum({ where: { id: thing.id } });
  const { object: updated } = await actions.get({ id: thing.id });
  expect(updated.enum).toEqual("TypeTwo");
});

test("enum set attribute - set to null - is null", async () => {
  const { object: thing } = await actions.create({});
  await actions.updateNullEnum({ where: { id: thing.id } });
  const { object: updated } = await actions.get({ id: thing.id });
  expect(updated.enum).toEqual(null);
});

test("enum set attribute from explicit input - set to TypeTwo - is TypeTwo", async () => {
  const { object: thing } = await actions.create({});
  const { object: thingUpdated } = await actions.updateEnumFromExplicitInput(
    { where: { id: thing.id }, values: { explEnum: "TypeTwo" } }
  );
  expect(thingUpdated.enum).toEqual("TypeTwo");
});

test("enum set attribute from implicit input - set to TypeTwo - is TypeTwo", async () => {
  const { object: thing } = await actions.create({});
  const { object: thingUpdated } = await actions.updateEnumFromImplicitInput(
    { where: { id: thing.id }, values: { otherEnum: "TypeTwo" } }
  );
  expect(thingUpdated.enum).toEqual("TypeTwo");
  expect(thingUpdated.otherEnum).toEqual("TypeTwo");
});
