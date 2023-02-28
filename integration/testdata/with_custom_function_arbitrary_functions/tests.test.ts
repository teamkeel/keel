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

  const result = await actions.countName({ name: "Keelson" });
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

  const result = await actions.countNameAdvanced({
    startsWith: "Kee",
    contains: "e",
    endsWith: "r",
  });
  expect(result.count).toEqual(1);
});

test("arbitrary write function with inline inputs", async () => {
  const result1 = await actions.createAndCount({ name: "Keelson" });
  expect(result1.count).toEqual(1);

  const result2 = await actions.createAndCount({ name: "Keelson" });
  expect(result2.count).toEqual(2);
});

test("arbitrary write function with message input", async () => {
  const result1 = await actions.createManyAndCount({
    names: ["Keelson", "Weaveton"],
  });
  expect(result1.count).toEqual(2);

  const result2 = await actions.createManyAndCount({
    names: ["Keelson", "Weaveton"],
  });
  expect(result2.count).toEqual(4);
});
