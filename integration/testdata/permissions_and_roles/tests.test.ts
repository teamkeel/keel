import { test, expect, beforeEach } from "vitest";
import { actions, resetDatabase } from "@teamkeel/testing";
import { models } from "@teamkeel/sdk";

beforeEach(resetDatabase);

test("permission set on model level for create op - matching title - is authorized", async () => {
  await expect(
    actions.create({ title: "hello", views: 0 })
  ).resolves.toMatchObject({
    title: "hello",
  });
});

test("permission set on model level for create op - not matching - is not authorized", async () => {
  await expect(
    actions.create({ title: "goodbye", views: 0 })
  ).toHaveAuthorizationError();
});

test("ORed permissions set on model level for get op - matching title - is authorized", async () => {
  const post = await actions.create({
    title: "hello",
    views: 0,
  });

  const p = await actions.get({ id: post.id });
  expect(p).toEqual(post);
});

test("ORed permissions set on model level for get op - matching title and views - is authorized", async () => {
  const post = await actions.create({ title: "hello", views: 5 });

  const p = await actions.get({ id: post.id });
  expect(p).toEqual(post);
});

test("ORed permissions set on model level for get op - none matching - is not authorized", async () => {
  const post = await actions.create({ title: "hello", views: 500 });

  await actions.update({
    where: { id: post.id },
    values: { title: "goodbye" },
  });

  await expect(actions.get({ id: post.id })).toHaveAuthorizationError();
});

test("no permissions set on model level for delete op - can delete - is authorized", async () => {
  const post = await actions.create({ title: "hello", views: 500 });

  await expect(actions.delete({ id: post.id })).resolves.toEqual(post.id);
});

test("text literal comparisons - all expressions fail - is not authorized", async () => {
  await expect(
    actions.textsFailedExpressions({ title: "hello" })
  ).toHaveAuthorizationError();
});

test("number literal comparisons - all expressions fail - is not authorized", async () => {
  await expect(
    actions.numbersFailedExpressions({ views: 2 })
  ).toHaveAuthorizationError();
});

test("boolean literal comparisons - all expressions fail - is not authorized", async () => {
  await expect(
    actions.booleansFailedExpressions({
      isActive: false,
    })
  ).toHaveAuthorizationError();
});

test("enum literal comparisons - all expressions fail - is not authorized", async () => {
  await expect(actions.enumFailedExpressions()).toHaveAuthorizationError();
});

test("permission role email is authorized", async () => {
  const { token } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "verified@agency.org",
      password: "1234",
    },
  });

  await models.identity.update(
    {
      email: "verified@agency.org",
      issuer: "keel",
    },
    {
      emailVerified: true,
    }
  );

  await expect(
    actions
      .withAuthToken(token)
      .createUsingRole({ title: "nothing special about this title" })
  ).resolves.toMatchObject({
    title: "nothing special about this title",
  });
});

test("permission role email is authorized but not verified", async () => {
  const { token } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "notVerified@agency.org",
      password: "1234",
    },
  });

  await expect(
    actions
      .withAuthToken(token)
      .createUsingRole({ title: "nothing special about this title" })
  ).toHaveAuthorizationError();
});

test("permission role wrong email is not authorized", async () => {
  const { token } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "editorFred42@agency.org",
      password: "1234",
    },
  });

  await expect(
    actions
      .withAuthToken(token)
      .createUsingRole({ title: "nothing special about this title" })
  ).toHaveAuthorizationError();
});

test("permission role domain is authorized", async () => {
  const { token } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "john.witherow@times.co.uk",
      password: "1234",
    },
  });

  await models.identity.update(
    {
      email: "john.witherow@times.co.uk",
      issuer: "keel",
    },
    {
      emailVerified: true,
    }
  );

  await expect(
    actions
      .withAuthToken(token)
      .createUsingRole({ title: "nothing special about this title" })
  ).resolves.toMatchObject({
    title: "nothing special about this title",
  });
});

test("permission role wrong domain is not authorized", async () => {
  const { token } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "jon.sargent@bbc.co.uk",
      password: "1234",
    },
  });

  await expect(
    actions
      .withAuthToken(token)
      .createUsingRole({ title: "nothing special about this title" })
  ).toHaveAuthorizationError();
});

// Regression test for: https://linear.app/keel/issue/RUN-179/role-based-permission-bug-fix
test("permission from unauthorized email is denied when model has only role-based permissions", async () => {
  const { token } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "imposter@gmail.com",
      password: "1234",
    },
  });

  await expect(
    actions.withAuthToken(token).doProcedure({ name: "frontal lobotomy" })
  ).toHaveAuthorizationError();
});
