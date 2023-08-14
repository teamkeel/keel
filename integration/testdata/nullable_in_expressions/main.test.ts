import { test, expect, beforeEach } from "vitest";
import { actions, resetDatabase, models } from "@teamkeel/testing";
import { Status } from "@teamkeel/sdk";

beforeEach(resetDatabase);

test("create operation - defaults", async () => {
  const person = await actions.createPersonWithDefaults();
  expect(person.name).toEqual("no name");
  expect(person.status).toEqual(Status.Fired);
});

test("create operation - set to null", async () => {
  const person = await actions.createPerson();
  expect(person.name).toBeNull();
  expect(person.status).toBeNull();
});

test("update operation - set to null", async () => {
  const { id } = await models.person.create({
    name: "Arnold",
    status: Status.Fired,
  });

  const person = await actions.updatePerson({ where: { id: id } });
  expect(person.name).toBeNull();
  expect(person.status).toBeNull();
});

test("list operation - filter by null", async () => {
  await models.person.create({ name: "Arnold", status: Status.Employed });
  await models.person.create({ name: "Bob", status: Status.Retrenched });

  let persons = await actions.uninitialesedPersons();
  expect(persons.results).toHaveLength(0);

  persons = await actions.listPersons();
  expect(persons.results).toHaveLength(2);

  await models.person.create({ name: null, status: Status.Retrenched });

  persons = await actions.uninitialesedPersons();
  expect(persons.results).toHaveLength(1);

  persons = await actions.listPersons();
  expect(persons.results).toHaveLength(2);

  await models.person.create({ name: "Dave", status: null });

  persons = await actions.uninitialesedPersons();
  expect(persons.results).toHaveLength(2);

  persons = await actions.listPersons();
  expect(persons.results).toHaveLength(2);

  await models.person.create({ name: null, status: null });

  persons = await actions.uninitialesedPersons();
  expect(persons.results).toHaveLength(3);

  persons = await actions.listPersons();
  expect(persons.results).toHaveLength(2);
});
