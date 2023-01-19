import { actions, models } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

test("creating a person", async () => {
  const person = await actions.createPerson({
    name: "foo",
    gender: "female",
    niNumber: "282",
  });

  expect(person.name).toEqual("foo");
});

test("fetching a person by id", async () => {
  const person = await models.person.create({
    name: "bar",
    gender: "male",
    niNumber: "123",
  });
  const fetchedPerson = await actions.getPerson({ id: person.id });

  expect(person.id).toEqual(fetchedPerson!.id);
});

test("fetching person by additional unique field (not PK)", async () => {
  const person = await models.person.create({
    name: "bar",
    gender: "male",
    niNumber: "333",
  });

  const fetchedPerson = await actions.getPersonByNINumber({ niNumber: "333" });

  expect(person.id).toEqual(fetchedPerson!.id);
});

test("listing", async () => {
  await models.person.create({ name: "fred", gender: "male", niNumber: "000" });
  const x11 = await models.person.create({
    name: "X11",
    gender: "alien",
    niNumber: "920",
  });
  const x22 = await models.person.create({
    name: "X22",
    gender: "alien",
    niNumber: "902",
  });

  const resp = await actions.listPeople({
    where: {
      gender: "alien",
    },
  });

  const alienNames = resp.results.map((a) => a.name);

  expect(alienNames).toEqual([x11.name, x22.name]);
});

test("deletion", async () => {
  const person = await models.person.create({
    name: "fred",
    gender: "male",
    niNumber: "678",
  });

  const deletedId = await actions.deletePerson({ id: person.id });

  expect(deletedId).toEqual(person.id);
});

test("updating", async () => {
  const person = await models.person.create({
    name: "fred",
    gender: "male",
    niNumber: "678",
  });

  const updatedPerson = await actions.updatePerson({
    where: { id: person.id },
    values: { name: "paul", gender: "non-binary", niNumber: "789" },
  });

  expect(updatedPerson.name).toEqual("paul");
  expect(updatedPerson.gender).toEqual("non-binary");
  expect(updatedPerson.niNumber).toEqual("789");
  expect(updatedPerson.id).toEqual(person.id);
});
