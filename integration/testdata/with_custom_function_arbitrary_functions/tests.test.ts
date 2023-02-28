import { actions, models, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

test("arbitrary read function with inline inputs", async () => {
  await models.person.create({
    name: "Keelson",
  });
  await models.person.create({
    name: "Weaveton",
  });
  await models.person.create({
    name: "Keeler",
  });

  var result = await actions.countName({ name: "Keelson" });
  expect(result.count).toEqual(1);
});

test("arbitrary read function with message input", async () => {
  await models.person.create({
    name: "Keelson",
  });
  await models.person.create({
    name: "Weaveton",
  });
  await models.person.create({
    name: "Keeler",
  });

  var result = await actions.countNameAdvanced({
    startsWith: "Kee",
    contains: "e",
    endsWith: "r",
  });
  expect(result.count).toEqual(1);
});

test("arbitrary write function with inline inputs", async () => {
  var result = await actions.createAndCount({ name: "Keelson" });
  expect(result.count).toEqual(1);

  result = await actions.createAndCount({ name: "Keelson" });
  expect(result.count).toEqual(2);
});

test("arbitrary write function with message input", async () => {
  var result = await actions.createManyAndCount({
    names: ["Keelson", "Weaveton"],
  });
  expect(result.count).toEqual(2);

  var result = await actions.createManyAndCount({
    names: ["Keelson", "Weaveton"],
  });
  expect(result.count).toEqual(4);
});
