import { actions, models, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

test("custom identity claims", async () => {
  const identity = await models.identity.create({
    email: "keel@keel.xyz",
    issuer: "google|123",
    teamId: "abc",
    myField1: "f1",
    myField2: "f2",
  });

  expect(identity.teamId).toEqual("abc");
  expect(identity.myField1).toEqual("f1");
  expect(identity.myField2).toEqual("f2");
});

test("custom identity claims - empty", async () => {
  const identity = await models.identity.create({
    email: "keel@keel.xyz",
    issuer: "google|123",
  });

  expect(identity.teamId).toBeNull();
  expect(identity.myField1).toBeNull();
  expect(identity.myField2).toBeNull();
});

test("custom identity claims - duplicate", async () => {
  await models.identity.create({
    email: "keel@keel.xyz",
    issuer: "google|123",
    teamId: "abc",
    myField1: "f1",
    myField2: "f2",
  });
  await models.identity.create({
    email: "keel@keel.xyz",
    issuer: "google|123",
    teamId: "mnu",
    myField1: "f1",
    myField2: "f2",
  });
  await models.identity.create({
    email: "keel@keel.xyz",
    issuer: "google|987",
    teamId: "mnu",
    myField1: "f1",
    myField2: "f2",
  });
  await models.identity.create({
    email: "keel@keel.xyz",
    issuer: "google|123",
    teamId: "xyz",
    myField1: "f1",
    myField2: "f2",
  });

  await expect(
    models.identity.create({
      email: "keel@keel.xyz",
      issuer: "google|987",
      teamId: "mnu",
      myField1: "f1",
      myField2: "f2",
    })
  ).toHaveError({});
});
