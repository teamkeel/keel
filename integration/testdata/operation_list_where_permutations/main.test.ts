import { test, expect, actions, Thing, logger } from "@teamkeel/testing";
import { LogLevel } from "@teamkeel/sdk";

test("equalsGettingStarted", async () => {
  await Thing.create({ title: "History of Art" });
  await Thing.create({ title: "History of Cars" });

  let resp;

  // The operations follow a naming convention.
  // E.g. eqTextFieldToInput should be read as:
  //
  // "using the Equals operator, on a field/input of type Text, with the LHS of
  // the expression being a reference to a field, and the RHS being a reference
  // to an Input"
  //
  // The other implied codes being Lit(eral) and null.
  //
  resp = await actions.eqTextFieldToInp({
    where: { whereArg: "History of Art" },
  });
  expect(resp.collection.length).toEqual(1);

  resp = await actions.eqTextFieldToField({});
  expect(resp.collection.length).toEqual(2);

  resp = await actions.eqTextFieldToLit({});
  expect(resp.collection.length).toEqual(1);

  resp = await actions.eqTextFieldToNil({});
  expect(resp.collection.length).toEqual(0);
});

test("equalsSwapLHSWithRHS", async () => {
  await Thing.create({ title: "History of Art" });
  await Thing.create({ title: "History of Cars" });

  let resp;

  resp = await actions.eqTextLitToField({});
  expect(resp.collection.length).toEqual(1);
});

test("notEqualSample", async () => {
  await Thing.create({ title: "History of Art" });
  await Thing.create({ title: "History of Cars" });

  let resp;

  resp = await actions.notEqTextFieldToLit({});
  expect(resp.collection.length).toEqual(1);
});

test("equalsWithNumberField", async () => {
  await Thing.create({ length: 41 });
  await Thing.create({ length: 42 });

  let resp;

  resp = await actions.eqNumberFieldToLit({});
  expect(resp.collection.length).toEqual(1);
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
