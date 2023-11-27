import { actions, models, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

test("create action with id input", async () => {
    await models.company.create({ id: "e1", name: "Keel" });

    const person = await actions.createPerson({ 
      id: "p1", 
      name: "Keelson",
      employer: {
        id: "e1"
      }  
    });

    expect(person.id).toEqual("p1");
    expect(person.employerId).toEqual("e1");

    const getPerson = await actions.getPerson({ id: "p1"});
    expect(getPerson?.id).toEqual("p1");
    expect(getPerson?.employerId).toEqual("e1");
});

test("create action with only id input", async () => {
  const person = await actions.createPersonOnlyId({ 
    id: "p1", 
  });

  expect(person.id).toEqual("p1");
  expect(person.employerId).toBeNull();

  const getPerson = await actions.getPerson({ id: "p1"});
  expect(getPerson?.id).toEqual("p1");
  expect(getPerson?.employerId).toBeNull();
});

test("create action with only id in @set", async () => {
  const person = await actions.createPersonOnlyIdWithSet({ 
    personId: "p1", 
  });

  expect(person.id).toEqual("p1");
  expect(person.employerId).toBeNull();

  const getPerson = await actions.getPerson({ id: "p1"});
  expect(getPerson?.id).toEqual("p1");
  expect(getPerson?.employerId).toBeNull();
});

test("create action with id in @set", async () => {
  await models.company.create({ id: "e1", name: "Keel" });

  const person = await actions.createPersonUsingSet({ 
    personId: "p1", 
    name: "Keelson",
    companyId: "e1"
  });

  expect(person.id).toEqual("p1");
  expect(person.employerId).toEqual("e1");

  const getPerson = await actions.getPerson({ id: "p1"});
  expect(getPerson?.id).toEqual("p1");
  expect(getPerson?.employerId).toEqual("e1");
});


test("create with nested company 1:M", async () => {
  const person = await actions.createWithEmployer({ 
    id: "p1", 
    name: "Keelson",
    employer: {
      id: "e1",
      name: "Keel"
    }
  });

  expect(person.id).toEqual("p1");
  expect(person.employerId).toEqual("e1");

  const getPerson = await actions.getPerson({ id: "p1"});
  expect(getPerson?.id).toEqual("p1");
  expect(getPerson?.employerId).toEqual("e1");

  const company = await models.company.findOne({ id: "e1" });
  expect(company?.id).toEqual("e1");
  expect(company?.name).toEqual("Keel");
});

test("create with nested company 1:M using @set", async () => {
  const person = await actions.createWithEmployerUsingSetId({ 
    id: "p1", 
    name: "Keelson",
    employerId: "e1",
    employer: {
      name: "Keel"
    }
  });

  expect(person.id).toEqual("p1");
  expect(person.employerId).toEqual("e1");

  const getPerson = await actions.getPerson({ id: "p1"});
  expect(getPerson?.id).toEqual("p1");
  expect(getPerson?.employerId).toEqual("e1");

  const company = await models.company.findOne({ id: "e1" });
  expect(company?.id).toEqual("e1");
  expect(company?.name).toEqual("Keel");
});

test("create with nested passport 1:1 with id", async () => {
  const person = await actions.createWithPassport({ 
    name: "Keelson",
    passport: {
      id: "pass1",
      number: "851119"
    }
  });

  expect(person.passportId).toEqual("pass1");
  expect(person.id).not.toBeNull();
  expect(person.id).not.toEqual("");

  const getPerson = await actions.getPerson({ id: person.id});
  expect(getPerson?.id).toEqual(person.id);
  expect(getPerson?.passportId).toEqual("pass1");

  const company = await models.passport.findOne({ id: "pass1" });
  expect(company?.id).toEqual("pass1");
  expect(company?.number).toEqual("851119");
});

test("create with nested passport 1:1 with id in @set", async () => {
  const person = await actions.createWithPasswordUsingSetId({ 
    name: "Keelson",
    passportId: "pass1",
    passport: {
      number: "851119"
    }
  });

  expect(person.passportId).toEqual("pass1");
  expect(person.id).not.toBeNull();
  expect(person.id).not.toEqual("");

  const getPerson = await actions.getPerson({ id: person.id});
  expect(getPerson?.id).toEqual(person.id);
  expect(getPerson?.passportId).toEqual("pass1");

  const company = await models.passport.findOne({ id: "pass1" });
  expect(company?.id).toEqual("pass1");
  expect(company?.number).toEqual("851119");
});

test("create with nested will 1:1 (inverse) with id", async () => {
  const person = await actions.createWithWill({ 
    id: "p1",
    name: "Keelson",
    will: {
      id: "w1",
      contents: "this is my will..."
    }
  });

  expect(person.id).toEqual("p1")

  const will = await models.will.findOne({ id: "w1" });
  expect(will?.id).toEqual("w1")
  expect(will?.contents).toEqual("this is my will...")
  expect(will?.personId).toEqual("p1")
});

test("create with nested will 1:1 (inverse) with id in @set", async () => {
  const person = await actions.createWithWillUsingSetId({ 
    id: "p1",
    name: "Keelson",
    willId: "w1",
    will: {
      contents: "this is my will..."
    }
  });

  expect(person.id).toEqual("p1")

  const will = await models.will.findOne({ id: "w1" });
  expect(will?.id).toEqual("w1")
  expect(will?.contents).toEqual("this is my will...")
  expect(will?.personId).toEqual("p1")
});

