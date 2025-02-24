import { actions, models, resetDatabase } from "@teamkeel/testing";
import { MyEnum } from "@teamkeel/sdk";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

test("array fields - create action with defaults", async () => {
  const thing = await actions.createThing();

  expect(thing.texts).toHaveLength(2);
  expect(thing.texts![0]).toEqual("Keel");
  expect(thing.texts![1]).toEqual("Weave");

  expect(thing.numbers).toHaveLength(3);
  expect(thing.numbers![0]).toEqual(1);
  expect(thing.numbers![1]).toEqual(2);
  expect(thing.numbers![2]).toEqual(3);

  expect(thing.booleans).toHaveLength(3);
  expect(thing.booleans![0]).toEqual(true);
  expect(thing.booleans![1]).toEqual(true);
  expect(thing.booleans![2]).toEqual(false);

  expect(thing.enums).toHaveLength(3);
  expect(thing.enums![0]).toEqual(MyEnum.One);
  expect(thing.enums![1]).toEqual(MyEnum.Two);
  expect(thing.enums![2]).toEqual(MyEnum.Three);

  expect(thing.enumsEmpty).toHaveLength(0);
  expect(thing.dates).toHaveLength(0);
  expect(thing.timestamps).toHaveLength(0);
  expect(thing.files).toHaveLength(0);
});
