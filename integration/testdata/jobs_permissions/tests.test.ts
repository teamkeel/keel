import { models, jobs, actions, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

async function jobRan(id) {
  const track = await models.trackJob.findOne({ id });
  return track!.didJobRun;
}

test("job - without identity - not permitted", async () => {
  const { id } = await models.trackJob.create({ didJobRun: false });

  await jobs.manualJob({ id });
  expect(await jobRan(id)).toBeFalsy();

  await jobs.manualJobIsAuthenticatedExpression({ id });

  expect(await jobRan(id)).toBeFalsy();

  await jobs.manualJobMultiRoles({ id });
  expect(await jobRan(id)).toBeFalsy();
});

test("job - invalid token - not authenticated", async () => {
  const { id } = await models.trackJob.create({ didJobRun: false });

  await jobs.withAuthToken("invalid").manualJobTrueExpression({ id });
  expect(await jobRan(id)).toBeFalsy();

  await jobs.withAuthToken("invalid").manualJob({ id });
  expect(await jobRan(id)).toBeFalsy();

  await jobs
    .withAuthToken("invalid")
    .manualJobIsAuthenticatedExpression({ id });
  expect(await jobRan(id)).toBeFalsy();

  await jobs.withAuthToken("invalid").manualJobMultiRoles({ id });
  expect(await jobRan(id)).toBeFalsy();
});

test("job - with identity, ctx.isAuthenticated - permitted", async () => {
  const { id } = await models.trackJob.create({ didJobRun: false });
  const identity = await models.identity.create({ email: "weave@gmail.com" });

  await jobs.withIdentity(identity).manualJobIsAuthenticatedExpression({ id });
  expect(await jobRan(id)).toBeTruthy();
});

test("job - with token, ctx.isAuthenticated - permitted", async () => {
  const { id } = await models.trackJob.create({ didJobRun: false });
  const identity = await models.identity.create({
    email: "weave@gmail.com",
  });

  await jobs.withIdentity(identity).manualJobIsAuthenticatedExpression({ id });
  expect(await jobRan(id)).toBeTruthy();
});

test("job - without identity, true expression permission - permitted", async () => {
  const { id } = await models.trackJob.create({ didJobRun: false });

  await jobs.manualJobTrueExpression({ id });
  expect(await jobRan(id)).toBeTruthy();
});

test("job - wrong domain - not permitted", async () => {
  const { id } = await models.trackJob.create({ didJobRun: false });
  const identity = await models.identity.create({
    email: "weave@gmail.com",
    emailVerified: true,
  });

  await jobs.withIdentity(identity).manualJob({ id });
  expect(await jobRan(id)).toBeFalsy();

  await jobs.withIdentity(identity).manualJobMultiRoles({ id });
  expect(await jobRan(id)).toBeFalsy();
});

test("job - authorised domain - permitted", async () => {
  const { id } = await models.trackJob.create({ didJobRun: false });
  const identity = await models.identity.create({
    email: "keel@keel.so",
    emailVerified: true,
  });

  await jobs.withIdentity(identity).manualJob({ id });
  expect(await jobRan(id)).toBeTruthy();
});

test("job - wrong authorised domain - not permitted", async () => {
  const { id } = await models.trackJob.create({ didJobRun: false });
  const identity = await models.identity.create({
    email: "keel@keel.dev",
    emailVerified: true,
  });

  await jobs.withIdentity(identity).manualJob({ id });
  expect(await jobRan(id)).toBeFalsy();
});

test("job - multi domains, authorised domain - permitted", async () => {
  const { id } = await models.trackJob.create({ didJobRun: false });
  const identity = await models.identity.create({
    email: "keel@keel.so",
    emailVerified: true,
  });

  await jobs.withIdentity(identity).manualJobMultiRoles({ id });
  expect(await jobRan(id)).toBeTruthy();
});

test("job - true expression - permitted", async () => {
  const { id } = await models.trackJob.create({ didJobRun: false });

  await jobs.manualJobTrueExpression({ id });
  expect(await jobRan(id)).toBeTruthy();
});

test("job - env var expression - permitted", async () => {
  const { id } = await models.trackJob.create({ didJobRun: false });

  await jobs.manualJobEnvExpression({ id });
  expect(await jobRan(id)).toBeTruthy();
});

test("job - env var expression fail - not permitted", async () => {
  const { id } = await models.trackJob.create({ didJobRun: false });

  await jobs.manualJobEnvExpression2({ id });
  expect(await jobRan(id)).toBeFalsy();
});

test("job - multiple permissions - not permitted", async () => {
  const { id } = await models.trackJob.create({ didJobRun: false });
  const identity = await models.identity.create({
    email: "bob@bob.com",
    emailVerified: true,
  });

  await jobs.withIdentity(identity).manualJobMultiPermission({ id });
  expect(await jobRan(id)).toBeFalsy();
});

test("job - multiple permissions - permitted", async () => {
  const { id } = await models.trackJob.create({ didJobRun: false });
  const identity = await models.identity.create({
    email: "keelson@keel.so",
    emailVerified: true,
  });

  await jobs.withIdentity(identity).manualJobMultiPermission({ id });
  expect(await jobRan(id)).toBeTruthy();
});

test("job - allowed in job code - permitted", async () => {
  const { id } = await models.trackJob.create({ didJobRun: false });
  const identity = await models.identity.create({
    email: "keel@keel.so",
    emailVerified: true,
  });

  await jobs
    .withIdentity(identity)
    .manualJobDeniedInCode({ id, denyIt: false });
  expect(await jobRan(id)).toBeTruthy();
});

test("job - denied in job code - not permitted without rollback transaction", async () => {
  const { id } = await models.trackJob.create({ didJobRun: false });
  const identity = await models.identity.create({
    email: "keel@keel.so",
    emailVerified: true,
  });

  await jobs.withIdentity(identity).manualJobDeniedInCode({ id, denyIt: true });

  // This would be false if a transaction rolled back.
  expect(await jobRan(id)).toBeTruthy();
});

test("job - exception - internal error without rollback transaction", async () => {
  const { id } = await models.trackJob.create({ didJobRun: false });
  const identity = await models.identity.create({
    email: "keel@keel.so",
    emailVerified: true,
  });

  await jobs.withIdentity(identity).manualJobWithException({ id });

  // This would be false if a transaction rolled back.
  expect(await jobRan(id)).toBeTruthy();
});

test("scheduled job - without identity - permitted", async () => {
  const { id } = await models.trackJob.create({
    id: "12345",
    didJobRun: false,
  });

  await jobs.scheduledWithoutPermissions({ scheduled: true });
  expect(await jobRan(id)).toBeTruthy();
});

test("scheduled job - with identity - permitted", async () => {
  const identity = await models.identity.create({ email: "keel@keel.so" });

  const { id } = await models.trackJob.create({
    id: "12345",
    didJobRun: false,
  });

  await jobs
    .withIdentity(identity)
    .scheduledWithoutPermissions({ scheduled: true });
  expect(await jobRan(id)).toBeTruthy();
});

test("scheduled job - invalid token - authentication failed", async () => {
  const { id } = await models.trackJob.create({
    id: "12345",
    didJobRun: false,
  });

  await jobs
    .withAuthToken("invalid")
    .scheduledWithoutPermissions({ scheduled: true });
  expect(await jobRan(id)).toBeFalsy();
});
