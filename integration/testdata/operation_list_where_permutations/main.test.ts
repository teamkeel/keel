import { test, expect, beforeEach } from "vitest";
import { actions, resetDatabase, models } from "@teamkeel/testing";

beforeEach(resetDatabase);

test("equalsGettingStarted", async () => {
  await models.thing.create({ title: "History of Art" });
  await models.thing.create({ title: "History of Cars" });

  // The actions follow a naming convention.
  // E.g. eqTextFieldToInput should be read as:
  //
  // "using the Equals operator, on a field/input of type Text, with the LHS of
  // the expression being a reference to a field, and the RHS being a reference
  // to an Input"
  //
  // The other implied codes being Lit(eral) and null.
  //
  let resp = await actions.eqTextFieldToInp({
    where: { whereArg: "History of Art" },
  });
  expect(resp.results.length).toEqual(1);

  resp = await actions.eqTextFieldToField();
  expect(resp.results.length).toEqual(2);

  resp = await actions.eqTextFieldToLit();
  expect(resp.results.length).toEqual(1);

  resp = await actions.eqTextFieldToNil();
  expect(resp.results.length).toEqual(0);
});

test("equalsSwapLHSWithRHS", async () => {
  await models.thing.create({ title: "History of Art" });
  await models.thing.create({ title: "History of Cars" });

  const resp = await actions.eqTextLitToField();
  expect(resp.results.length).toEqual(1);
});

test("notEqualSample", async () => {
  await models.thing.create({ title: "History of Art" });
  await models.thing.create({ title: "History of Cars" });

  const resp = await actions.notEqTextFieldToLit();
  expect(resp.results.length).toEqual(1);
});

test("equalsWithNumberField", async () => {
  await models.thing.create({ length: 41 });
  await models.thing.create({ length: 42 });

  const resp = await actions.eqNumberFieldToLit();
  expect(resp.results.length).toEqual(1);
});

/*

  all types:
  - text
  - number
  - bool
  - enum
  - field
  - null  
  */

test("inTextFieldToLit", async () => {
  const matchingModel = await models.thing.create({ title: "title1" });
  await models.thing.create({ title: "title2" });
  await models.thing.create({ title: "title3" });

  const resp = await actions.inTextFieldToLit();
  expect(resp.results.length).toEqual(1);
  expect(resp.results[0].id).toEqual(matchingModel.id);
});

test("notInTextFieldToLit", async () => {
  const matchingModel = await models.thing.create({ title: "title1" });
  await models.thing.create({ title: "title2" });
  await models.thing.create({ title: "title3" });

  const resp = await actions.notInTextFieldToLit();
  expect(resp.results.length).toEqual(1);
  expect(resp.results[0].id).toEqual(matchingModel.id);
});
