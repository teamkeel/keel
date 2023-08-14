import { test, expect, beforeEach } from "vitest";
import { actions, resetDatabase, models } from "@teamkeel/testing";
import { Status } from "@teamkeel/sdk";

beforeEach(resetDatabase);

test("create action - null values", async () => {
  const identity = await models.identity.create({
    email: "dave@keel.xyz",
    password: "123",
  });

  const createdPerson = await actions.createPerson({
    name: "Arnold",
    preferredName: null,
    employmentStatus: null,
    employer: { id: null },
  });

  // Test the response from createPerson
  expect(createdPerson.name).toEqual("Arnold");
  expect(createdPerson.preferredName).toBeNull();
  expect(createdPerson.employmentStatus).toBeNull();
  expect(createdPerson.employerId).toBeNull();

  const getPerson = await actions.getPerson({ id: createdPerson.id });

  // Test the response from getPerson
  expect(getPerson!.name).toEqual("Arnold");
  expect(getPerson!.preferredName).toBeNull();
  expect(getPerson!.employmentStatus).toBeNull();
  expect(getPerson!.employerId).toBeNull();
});

test("nested create action - null values", async () => {
  const identity = await models.identity.create({
    email: "dave@keel.xyz",
    password: "123",
  });

  const createdPerson = await actions.createPersonAndEmployer({
    name: "Arnold",
    employer: { tradingAs: null },
  });

  const getPerson = await actions.getPerson({ id: createdPerson.id });

  const getCompany = await models.company.findOne({
    id: getPerson!.employerId!,
  });

  // Test the response from getPerson
  expect(getPerson!.name).toEqual("Arnold");
  expect(getCompany!.tradingAs).toBeNull();
});

test("update action - null values", async () => {
  const identity = await models.identity.create({
    email: "dave@keel.xyz",
    password: "123",
  });
  const company = await models.company.create({ tradingAs: "Hollywood" });
  const person = await models.person.create({
    name: "Arnold",
    preferredName: "Arnie",
    employmentStatus: Status.Employed,
    employerId: company.id,
  });

  let getPerson = await actions.getPerson({ id: person.id });

  // Test the response from getPerson
  expect(getPerson!.name).toEqual("Arnold");
  expect(getPerson!.preferredName).toEqual("Arnie");
  expect(getPerson!.employmentStatus).toEqual(Status.Employed);
  expect(getPerson!.employerId).toEqual(company.id);

  const updatedPerson = await actions.updatePerson({
    where: { id: getPerson!.id },
    values: {
      preferredName: null,
      employmentStatus: null,
      employer: { id: null },
    },
  });

  // Test the response from updatedPerson
  expect(updatedPerson!.name).toEqual("Arnold");
  expect(updatedPerson!.preferredName).toBeNull();
  expect(updatedPerson!.employmentStatus).toBeNull();
  expect(updatedPerson!.employerId).toBeNull();
});

test("list action - null values", async () => {
  const identity = await models.identity.create({
    email: "dave@keel.xyz",
    password: "123",
  });

  const company = await models.company.create({ tradingAs: "Hollywood" });
  const companyWithNull = await models.company.create({ tradingAs: null });
  await models.person.create({
    name: "Arnold With All Data",
    preferredName: "Arnie",
    employmentStatus: Status.Employed,
    employerId: company.id,
  });
  await models.person.create({
    name: "Arnold Without Name",
    preferredName: null,
    employmentStatus: Status.Employed,
    employerId: company.id,
  });
  await models.person.create({
    name: "Arnold Without Status",
    preferredName: "Arnie",
    employmentStatus: null,
    employerId: company.id,
  });
  await models.person.create({
    name: "Arnold Without Employer",
    preferredName: "Arnie",
    employmentStatus: Status.Employed,
    employerId: companyWithNull.id,
  });
  await models.person.create({
    name: "Arnold No Data",
    preferredName: null,
    employmentStatus: null,
    employerId: companyWithNull.id,
  });

  let { results: resultsAllData } = await actions.listPersons({
    where: {
      preferredName: { notEquals: null },
      employmentStatus: { notEquals: null },
      employer: { tradingAs: { notEquals: null } },
    },
  });
  expect(resultsAllData).length(1);
  expect(resultsAllData[0].name).toEqual("Arnold With All Data");

  let { results: resultsNoData } = await actions.listPersons({
    where: {
      preferredName: { equals: null },
      employmentStatus: { equals: null },
      employer: { tradingAs: { equals: null } },
    },
  });
  expect(resultsNoData).length(1);
  expect(resultsNoData[0].name).toEqual("Arnold No Data");

  let { results: resultsNullPreferredName } = await actions.listPersons({
    where: {
      preferredName: { equals: null },
      employmentStatus: { notEquals: null },
      employer: { tradingAs: { notEquals: null } },
    },
  });
  expect(resultsNullPreferredName).length(1);
  expect(resultsNullPreferredName[0].name).toEqual("Arnold Without Name");

  let { results: resultsNullStatus } = await actions.listPersons({
    where: {
      preferredName: { notEquals: null },
      employmentStatus: { equals: null },
      employer: { tradingAs: { notEquals: null } },
    },
  });
  expect(resultsNullStatus).length(1);
  expect(resultsNullStatus[0].name).toEqual("Arnold Without Status");

  let { results: resultsNullEmployerTradingAs } = await actions.listPersons({
    where: {
      preferredName: { notEquals: null },
      employmentStatus: { notEquals: null },
      employer: { tradingAs: { equals: null } },
    },
  });
  expect(resultsNullEmployerTradingAs).length(1);
  expect(resultsNullEmployerTradingAs[0].name).toEqual(
    "Arnold Without Employer"
  );
});
