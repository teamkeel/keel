import { actions, resetDatabase, models, jobs } from "@teamkeel/testing";
import { beforeEach, expect, test } from "vitest";

beforeEach(resetDatabase);

test("events from action", async () => {
  await actions.createPerson({ name: "Keelson", email: "keelson@keel.so" });

  const persons = await models.person.findMany();

  expect(persons).toHaveLength(1);
  expect(persons[0].verifiedEmail).toBeTruthy();
  expect(persons[0].verifiedUpdate).toBeTruthy();
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
  await jobs.createRandomPersons();

  const persons = await models.person.findMany();

  expect(persons).toHaveLength(2);
  expect(persons[0].verifiedEmail).toBeTruthy();
  expect(persons[1].verifiedEmail).toBeTruthy();
  expect(persons[0].verifiedUpdate).toBeTruthy();
  expect(persons[1].verifiedUpdate).toBeTruthy();
});
