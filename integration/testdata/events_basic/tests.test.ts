import { actions, resetDatabase, models, jobs } from "@teamkeel/testing";
import { beforeEach, expect, test } from "vitest";

beforeEach(resetDatabase);

// test("events from action", async () => {
//   await actions.createPerson({ name: "Keelson", email: "keelson@keel.so" });
//   await actions.createPersonFn({ name: "Weaveton", email: "weaveton@keel.so" });

//   const persons = await models.person.findMany();

//   expect(persons).toHaveLength(2);
//   expect(persons[0].verifiedEmail).toBeTruthy();
//   expect(persons[1].verifiedEmail).toBeTruthy();
//   expect(persons[0].verifiedUpdate).toBeTruthy();
//   expect(persons[1].verifiedUpdate).toBeTruthy();
// });

// test("events from hook functions", async () => {
//   await actions.createPersonFn({ name: "Keelson", email: "keelson@keel.so" });
//   await actions.createPersonFn({ name: "Weaveton", email: "weaveton@keel.so" });

//   const persons = await models.person.findMany();

//   expect(persons).toHaveLength(2);
//   expect(persons[0].verifiedEmail).toBeTruthy();
//   expect(persons[1].verifiedEmail).toBeTruthy();
//   expect(persons[0].verifiedUpdate).toBeTruthy();
//   expect(persons[1].verifiedUpdate).toBeTruthy();
// });

// test("events from custom function", async () => {
//   const result = await actions.writeRandomPersons();
//   expect(result).toBeTruthy();

//   const persons = await models.person.findMany();

//   expect(persons).toHaveLength(2);
//   expect(persons[0].verifiedEmail).toBeTruthy();
//   expect(persons[1].verifiedEmail).toBeTruthy();
//   expect(persons[0].verifiedUpdate).toBeTruthy();
//   expect(persons[1].verifiedUpdate).toBeTruthy();
// });

// test("events from job", async () => {
//   await jobs.createRandomPersons({ raiseException: false });

//   const persons = await models.person.findMany();

//   expect(persons).toHaveLength(2);
//   expect(persons[0].verifiedEmail).toBeTruthy();
//   expect(persons[1].verifiedEmail).toBeTruthy();
//   expect(persons[0].verifiedUpdate).toBeTruthy();
//   expect(persons[1].verifiedUpdate).toBeTruthy();
// });

// test("events from failed hook function with rollback", async () => {
//   await expect(
//     actions.createPersonFn({ name: "", email: "keelson@keel.so" })
//   ).toHaveError({
//     code: "ERR_UNKNOWN",
//   });

//   const persons = await models.person.findMany();
//   expect(persons).toHaveLength(0);
// });

// test("events from failed job", async () => {
//   await jobs.createRandomPersons({ raiseException: true });

//   const persons = await models.person.findMany();

//   expect(persons).toHaveLength(2);
//   expect(persons[0].verifiedEmail).toBeTruthy();
//   expect(persons[1].verifiedEmail).toBeTruthy();
//   expect(persons[0].verifiedUpdate).toBeTruthy();
//   expect(persons[1].verifiedUpdate).toBeTruthy();
// });

// test("events from failed custom function with rollback", async () => {
//   await expect(
//     actions.writeRandomPersons({ raiseException: true })
//   ).toHaveError({
//     code: "ERR_UNKNOWN",
//   });

//   const persons = await models.person.findMany();
//   expect(persons).toHaveLength(0);
// });

// test("event previous data", async () => {
//   const t = await models.tracker.create({
//     views: 0,
//   });

//   const updated = await actions.updateViews({ where: { id: t.id }, values: { views: 1 } });

//   const get = await models.tracker.findOne({ id: t.id });

//   expect(get!.views).toEqual(1);
//   expect(get!.verifiedUpdate).toBeTruthy();

// });

test("event previous data bulk", async () => {
  await models.tracker.create({
    views: 3,
  });

  await models.tracker.create({
    views: 0,
  });

  await models.tracker.create({
    views: 10,
  });

  await models.tracker.create({
    views: 8,
  });

  await models.tracker.create({
    views: 2,
  });

  await models.tracker.create({
    views: 4,
  });

  await actions.updateTrackers({});
  await actions.updateTrackers({});
  await actions.updateTrackers({});
  await actions.updateTrackers({});
  await actions.updateTrackers({});

  const trackers = await models.tracker.findMany();

  for (const t of trackers) {
    expect(t.verifiedUpdate).toBeTruthy();
  }
});
