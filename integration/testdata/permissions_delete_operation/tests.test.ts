import { actions } from "@teamkeel/testing";
import { test, expect } from "vitest";

test("string permission on literal - matching value - is authorized", async () => {
  const post = await actions.createWithTitle({ title: "hello" });

  const deleted = await actions.deleteWithTextPermissionLiteral({
    id: post.id,
  });
  expect(deleted).toEqual(post.id);
});

test("string permission on literal - not matching value - is not authorized", async () => {
  const post = await actions.createWithTitle({
    title: "not hello",
  });

  await expect(
    actions.deleteWithTextPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("string permission on literal - not matching null value - is not authorized", async () => {
  const post = await actions.createWithTitle({ title: null });

  await expect(
    actions.deleteWithTextPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("number permission on literal - matching value - is authorized", async () => {
  const post = await actions.createWithViews({ views: 5 });

  const deleted = await actions.deleteWithNumberPermissionLiteral({
    id: post.id,
  });
  expect(deleted).toEqual(post.id);
});

test("number permission on literal - not matching value - is not authorized", async () => {
  const post = await actions.createWithViews({ views: 500 });

  await expect(
    actions.deleteWithNumberPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("number permission on literal - not matching null value - is not authorized", async () => {
  const post = await actions.createWithViews({ views: null });

  await expect(
    actions.deleteWithNumberPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("boolean permission on literal - matching value - is authorized", async () => {
  const post = await actions.createWithActive({ active: true });

  const deleted = await actions.deleteWithBooleanPermissionLiteral({
    id: post.id,
  });
  expect(deleted).toEqual(post.id);
});

test("boolean permission on literal - not matching value - is not authorized", async () => {
  const post = await actions.createWithActive({ active: false });

  await expect(
    actions.deleteWithBooleanPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("boolean permission on literal - not matching null value - is not authorized", async () => {
  const post = await actions.createWithActive({ active: null });

  await expect(
    actions.deleteWithBooleanPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("enum permission on literal - matching value - is authorized", async () => {
  const post = await actions.createWithEnum({ type: "Technical" });

  const deleted = await actions.deleteWithEnumPermissionLiteral({
    id: post.id,
  });
  expect(deleted).toEqual(post.id);
});

test("enum permission on literal - not matching value - is not authorized", async () => {
  const post = await actions.createWithEnum({ title: "Lifestyle" });

  await expect(
    actions.deleteWithEnumPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("enum permission on literal - not matching null value - is not authorized", async () => {
  const post = await actions.createWithEnum({ title: null });

  await expect(
    actions.deleteWithEnumPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("string permission on field - matching value - is authorized", async () => {
  const post = await actions.createWithTitle({ title: "hello" });

  const deleted = await actions.deleteWithTextPermissionOnField({
    id: post.id,
  });
  expect(deleted).toEqual(post.id);
});

test("string permission on field - not matching value - is not authorized", async () => {
  const post = await actions.createWithTitle({
    title: "not hello",
  });

  await expect(
    actions.deleteWithTextPermissionOnField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("string permission on field - not matching null value - is not authorized", async () => {
  const post = await actions.createWithTitle({ title: null });

  await expect(
    actions.deleteWithTextPermissionOnField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("number permission on field - matching value - is authorized", async () => {
  const post = await actions.createWithViews({ views: 5 });

  const deleted = await actions.deleteWithNumberPermissionOnField({
    id: post.id,
  });
  expect(deleted).toEqual(post.id);
});

test("number permission on field - not matching value - is not authorized", async () => {
  const post = await actions.createWithViews({ views: 500 });

  await expect(
    actions.deleteWithNumberPermissionOnField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("number permission on field - not matching null value - is not authorized", async () => {
  const post = await actions.createWithViews({ views: null });

  await expect(
    actions.deleteWithNumberPermissionOnField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("boolean permission on field - matching value - is authorized", async () => {
  const post = await actions.createWithActive({ active: true });

  const deleted = await actions.deleteWithBooleanPermissionOnField({
    id: post.id,
  });
  expect(deleted).toEqual(post.id);
});

test("boolean permission on field - not matching value - is not authorized", async () => {
  const post = await actions.createWithActive({ active: false });

  await expect(
    actions.deleteWithBooleanPermissionOnField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("boolean permission on field - not matching null value - is not authorized", async () => {
  const post = await actions.createWithActive({ active: null });

  await expect(
    actions.deleteWithBooleanPermissionOnField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("enum permission on field - matching value - is authorized", async () => {
  const post = await actions.createWithEnum({ type: "Technical" });

  const deleted = await actions.deleteWithEnumPermissionOnField({
    id: post.id,
  });
  expect(deleted).toEqual(post.id);
});

test("enum permission on field - not matching value - is not authorized", async () => {
  const post = await actions.createWithEnum({ type: "Lifestyle" });

  await expect(
    actions.deleteWithEnumPermissionOnField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("enum permission on field - not matching null value - is not authorized", async () => {
  const post = await actions.createWithEnum({ type: null });

  await expect(
    actions.deleteWithEnumPermissionOnField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("identity permission - correct identity in context - is authorized", async () => {
  const { token } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "user@keel.xyz",
      password: "1234",
    },
  });

  const post = await actions.withAuthToken(token).createWithIdentity({});

  const deleted = await actions
    .withAuthToken(token)
    .deleteWithRequiresSameIdentity({ id: post.id });

  expect(deleted).toEqual(post.id);
});

test("identity permission - incorrect identity in context - is not authorized", async () => {
  const { token } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "user1@keel.xyz",
      password: "1234",
    },
  });

  const { token: token2 } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "user2@keel.xyz",
      password: "1234",
    },
  });

  const post = await actions.withAuthToken(token).createWithIdentity({});

  await expect(
    actions
      .withAuthToken(token2)
      .deleteWithRequiresSameIdentity({ id: post.id })
  ).toHaveAuthorizationError();
});

test("identity permission - no identity in context - is not authorized", async () => {
  const { token } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "user@keel.xyz",
      password: "1234",
    },
  });

  const post = await actions.withAuthToken(token).createWithIdentity({});

  await expect(
    actions.deleteWithRequiresSameIdentity({ id: post.id })
  ).toHaveAuthorizationError();
});

test("true value permission - with unauthenticated identity - is authorized", async () => {
  const post = await actions.createWithActive({ active: true });

  const deleted = await actions.deleteWithTrueValuePermission({ id: post.id });
  expect(deleted).toEqual(post.id);
});

test("permission on implicit input - matching value - is authorized", async () => {
  const post = await actions.createWithTitle({ title: "hello" });

  const deleted = await actions.deleteWithPermissionOnImplicitInput({
    id: post.id,
    title: "hello",
  });
  expect(deleted).toEqual(post.id);
});

test("permission on implicit input - not matching value - is not authorized", async () => {
  const post = await actions.createWithTitle({ title: "hello" });

  await expect(
    actions.deleteWithPermissionOnImplicitInput({
      id: post.id,
      title: "goodbye",
    })
  ).toHaveAuthorizationError();
});

test("permission on implicit input - not matching null value - is not authorized", async () => {
  const post = await actions.createWithTitle({ title: "hello" });

  await expect(
    actions.deleteWithPermissionOnImplicitInput({
      id: post.id,
      title: null,
    })
  ).toHaveAuthorizationError();
});

test("permission on explicit input - not matching value - is not authorized", async () => {
  const post = await actions.createWithTitle({ title: "hello" });

  await expect(
    actions.deleteWithPermissionOnExplicitInput({
      id: post.id,
      explTitle: "goodbye",
    })
  ).toHaveAuthorizationError();
});

test("permission on explicit input - matching null value - is authorized", async () => {
  const post = await actions.createWithTitle({ title: null });

  const deleted = await actions.deleteWithPermissionOnExplicitInput({
    id: post.id,
    explTitle: null,
  });
  expect(deleted).toEqual(post.id);
});

test("permission on explicit input - not matching null value - is not authorized", async () => {
  const post = await actions.createWithTitle({ title: "hello" });

  await expect(
    actions.deleteWithPermissionOnExplicitInput({
      id: post.id,
      explTitle: null,
    })
  ).toHaveAuthorizationError();
});
