import { actions, resetDatabase, models, jobs } from "@teamkeel/testing";
import { beforeEach, expect, test } from "vitest";

beforeEach(resetDatabase);

test("events from action", async () => {
  await actions.createPerson({ name: "Keelson", email: "keelson@keel.so" });
  await actions.createPersonFn({ name: "Weaveton", email: "weaveton@keel.so" });

  const persons = await models.person.findMany();

  expect(persons).toHaveLength(2);
  expect(persons[0].verifiedEmail).toBeTruthy();
  expect(persons[1].verifiedEmail).toBeTruthy();
  expect(persons[0].verifiedUpdate).toBeTruthy();
  expect(persons[1].verifiedUpdate).toBeTruthy();
});

test("events from hook functions", async () => {
  await actions.createPersonFn({ name: "Keelson", email: "keelson@keel.so" });
  await actions.createPersonFn({ name: "Weaveton", email: "weaveton@keel.so" });

  const persons = await models.person.findMany();

  expect(persons).toHaveLength(2);
  expect(persons[0].verifiedEmail).toBeTruthy();
  expect(persons[1].verifiedEmail).toBeTruthy();
  expect(persons[0].verifiedUpdate).toBeTruthy();
  expect(persons[1].verifiedUpdate).toBeTruthy();
});

test("events from custom function", async () => {
  const result = await actions.writeRandomPersons();
  expect(result).toBeTruthy();

  const persons = await models.person.findMany();

  expect(persons).toHaveLength(2);
  expect(persons[0].verifiedEmail).toBeTruthy();
  expect(persons[1].verifiedEmail).toBeTruthy();
  expect(persons[0].verifiedUpdate).toBeTruthy();
  expect(persons[1].verifiedUpdate).toBeTruthy();
});

test("events from job", async () => {
  await jobs.createRandomPersons({ raiseException: false });

  const persons = await models.person.findMany();

  expect(persons).toHaveLength(2);
  expect(persons[0].verifiedEmail).toBeTruthy();
  expect(persons[1].verifiedEmail).toBeTruthy();
  expect(persons[0].verifiedUpdate).toBeTruthy();
  expect(persons[1].verifiedUpdate).toBeTruthy();
});

test("events from failed hook function with rollback", async () => {
  await expect(
    actions.createPersonFn({ name: "", email: "keelson@keel.so" })
  ).toHaveError({
    code: "ERR_INTERNAL",
  });

  const persons = await models.person.findMany();
  expect(persons).toHaveLength(0);
});

test("events from failed job", async () => {
  await expect(jobs.createRandomPersons({ raiseException: true })).toHaveError({
    code: "ERR_INTERNAL",
  });

  const persons = await models.person.findMany();

  expect(persons).toHaveLength(2);
  expect(persons[0].verifiedEmail).toBeTruthy();
  expect(persons[1].verifiedEmail).toBeTruthy();
  expect(persons[0].verifiedUpdate).toBeTruthy();
  expect(persons[1].verifiedUpdate).toBeTruthy();
});

test("events from failed custom function with rollback", async () => {
  await expect(
    actions.writeRandomPersons({ raiseException: true })
  ).toHaveError({
    code: "ERR_INTERNAL",
  });

  const persons = await models.person.findMany();
  expect(persons).toHaveLength(0);
});
