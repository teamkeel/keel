import { actions, models } from "@teamkeel/testing";
import { test, expect, beforeAll } from "vitest";
import { Status } from "@teamkeel/sdk";

beforeAll(async () => {
  await models.person.create({
    text: "matching",
    bool: true,
    enum: Status.Option1,
    number: 100,
  });

  await models.person.create({
    text: "unmatching",
    bool: false,
    enum: Status.Option2,
    number: 1,
  });

  await models.personOptionalFields.create({
    text: "matching",
    bool: true,
    enum: Status.Option1,
    number: 100,
  });

  await models.personOptionalFields.create({
    text: "unmatching",
    bool: false,
    enum: Status.Option2,
    number: 1,
  });
});

test("listing one with optional text input", async () => {
  let resp = await actions.listOptionalInputs({
    where: {
      text: { startsWith: "mat" },
    },
  });

  expect(resp.results.length).toEqual(1);
  expect(resp.results[0].text).toEqual("matching");
});

test("listing one with optional bool input", async () => {
  let resp = await actions.listOptionalInputs({
    where: {
      bool: { equals: true },
    },
  });

  expect(resp.results.length).toEqual(1);
  expect(resp.results[0].text).toEqual("matching");
});

test("listing one with optional enum input", async () => {
  let resp = await actions.listOptionalInputs({
    where: {
      enum: { oneOf: [Status.Option1] },
    },
  });

  expect(resp.results.length).toEqual(1);
  expect(resp.results[0].text).toEqual("matching");
});

test("listing one with optional enum input", async () => {
  let resp = await actions.listOptionalInputs({
    where: {
      number: { greaterThan: 99 },
    },
  });

  expect(resp.results.length).toEqual(1);
  expect(resp.results[0].text).toEqual("matching");
});

test("listing both with all optional fields input", async () => {
  let respBoth = await actions.listOptionalInputs({
    where: {
      text: { contains: "mat" },
      enum: { oneOf: [Status.Option1, Status.Option2] },
      number: { lessThan: 999999 },
    },
  });

  expect(respBoth.results.length).toEqual(2);
});

test("listing one with all optional fields input", async () => {
  let respMatching = await actions.listOptionalInputs({
    where: {
      text: { startsWith: "mat" },
      bool: { equals: true },
      enum: { oneOf: [Status.Option1] },
      number: { greaterThan: 99 },
    },
  });

  expect(respMatching.results.length).toEqual(1);
  expect(respMatching.results[0].text).toEqual("matching");
});

test("listing one with optional text input and optional field", async () => {
  let resp = await actions.listOptionalFieldsWithOptionalInputs({
    where: {
      text: { startsWith: "mat" },
    },
  });

  expect(resp.results.length).toEqual(1);
  expect(resp.results[0].text).toEqual("matching");
});

test("listing one with optional bool input and optional field", async () => {
  let resp = await actions.listOptionalFieldsWithOptionalInputs({
    where: {
      bool: { equals: true },
    },
  });

  expect(resp.results.length).toEqual(1);
  expect(resp.results[0].text).toEqual("matching");
});

test("listing one with optional enum input and optional field", async () => {
  let resp = await actions.listOptionalFieldsWithOptionalInputs({
    where: {
      enum: { oneOf: [Status.Option1] },
    },
  });

  expect(resp.results.length).toEqual(1);
  expect(resp.results[0].text).toEqual("matching");
});

test("listing one with optional enum input and optional field", async () => {
  let resp = await actions.listOptionalFieldsWithOptionalInputs({
    where: {
      number: { greaterThan: 99 },
    },
  });

  expect(resp.results.length).toEqual(1);
  expect(resp.results[0].text).toEqual("matching");
});

test("listing both with all optional inputs and optional field", async () => {
  let respBoth = await actions.listOptionalFieldsWithOptionalInputs({
    where: {
      text: { contains: "mat" },
      enum: { oneOf: [Status.Option1, Status.Option2] },
      number: { lessThan: 999999 },
    },
  });

  expect(respBoth.results.length).toEqual(2);
});

test("listing one with all optional input and optional field", async () => {
  let respMatching = await actions.listOptionalFieldsWithOptionalInputs({
    where: {
      text: { startsWith: "mat" },
      bool: { equals: true },
      enum: { oneOf: [Status.Option1] },
      number: { greaterThan: 99 },
    },
  });

  expect(respMatching.results.length).toEqual(1);
  expect(respMatching.results[0].text).toEqual("matching");
});

test("listing one with all required fields and required inputs", async () => {
  let respMatching = await actions.listRequiredInputs({
    where: {
      text: { startsWith: "mat" },
      bool: { equals: true },
      enum: { oneOf: [Status.Option1] },
      number: { greaterThan: 99 },
    },
  });

  expect(respMatching.results.length).toEqual(1);
  expect(respMatching.results[0].text).toEqual("matching");
});

test("listing one with all optional fields and required inputs", async () => {
  let respMatching = await actions.listOptionalFieldsWithRequiredInputs({
    where: {
      text: { startsWith: "mat" },
      bool: { equals: true },
      enum: { oneOf: [Status.Option1] },
      number: { greaterThan: 99 },
    },
  });

  expect(respMatching.results.length).toEqual(1);
  expect(respMatching.results[0].text).toEqual("matching");
});

test("listing with number one of", async () => {
  let resp = await actions.listOptionalInputs({
    where: {
      number: { oneOf: [100] },
    },
  });

  expect(resp.results.length).toEqual(1);
  expect(resp.results[0].text).toEqual("matching");
});
