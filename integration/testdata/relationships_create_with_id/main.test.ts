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

test("create action with id input - already exists", async () => {
  await models.company.create({ id: "e1", name: "Keel" });
  await models.person.create({ id: "p1" });

  await expect(
    actions.createPerson({ 
      id: "p1", 
      name: "Keelson",
      employer: {
        id: "e1"
      }  
    })
  ).toHaveError({
    code: "ERR_INVALID_INPUT",
    message: "the value for the unique field 'id' must be unique",
  });
});

test("update action with id input", async () => {
  await models.person.create({ id: "p1", name: "Keelson" });

  const person = await actions.updatePersonId({ 
    where: { id: "p1" },
    values: { id: "p2" }
  });

  expect(person.id).toEqual("p2");
});

test("update action with id in @set", async () => {
  await models.person.create({ id: "p1", name: "Keelson" });

  const person = await actions.updatePersonIdWithSet({ 
    where: { id: "p1" },
    values: { newId: "p2" }
  });

  expect(person.id).toEqual("p2");
});

test("update action with company id input", async () => {
  await models.company.create({ id: "e1", name: "Keel" });
  await models.company.create({ id: "e2", name: "Keel" });
  await models.person.create({ id: "p1", name: "Keelson", employerId: "e1" });

  const person = await actions.updatePersonCompanyId({ 
    where: { id: "p1" },
    values: { employer: { id: "e2" } }
  });

  expect(person.employerId).toEqual("e2");
});

test("update action with company id input - does not exist", async () => {
  await models.company.create({ id: "e1", name: "Keel" });
  await models.person.create({ id: "p1", name: "Keelson", employerId: "e1" });

 await expect( actions.updatePersonCompanyId({ 
    where: { id: "p1" },
    values: { employer: { id: "nope" } }
  })).toHaveError({
    code: "ERR_INVALID_INPUT",
    message: "the record referenced in field 'employerId' does not exist",
  });
});

test("update action with id input - already exists", async () => {
  await models.person.create({ id: "p1", name: "Keelson" });
  await models.person.create({ id: "p2", name: "Keeler" });


  await expect( actions.updatePersonId({ 
    where: { id: "p1" },
    values: { id: "p2" }
  })
  ).toHaveError({
    code: "ERR_INVALID_INPUT",
    message: "the value for the unique field 'id' must be unique",
  });
});

test("update action with id input - has foreign key references", async () => {

  await models.person.create({ id: "p1", name: "Keelson" });
  await models.will.create({ personId: "p1", contents: "I declare..." });

  await expect( actions.updatePersonId({ 
    where: { id: "p1" },
    values: { id: "p2" }
  })
  ).toHaveError({
    code: "ERR_INVALID_INPUT",
    message: "the record referenced in field 'id' does not exist",
  });
});

