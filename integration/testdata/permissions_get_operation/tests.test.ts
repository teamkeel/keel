import { actions } from "@teamkeel/testing";
import { test, expect } from "vitest";
import { PostType } from "@teamkeel/sdk";

test("string permission on literal - matching value - is authorized", async () => {
  const post = await actions.createWithText({ title: { value: "hello" } });

  const p = await actions.getWithTextPermissionLiteral({ id: post.id });
  expect(p!.id).toEqual(post.id);
});

test("string permission on literal - not matching value - is not authorized", async () => {
  const post = await actions.createWithText({ title: { value: "goodbye" } });

  await expect(
    actions.getWithTextPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("string permission on literal - null value - is not authorized", async () => {
  const post = await actions.createWithText({ title: { isNull: true } });

  await expect(
    actions.getWithTextPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("number permission on literal - matching value - is authorized", async () => {
  const post = await actions.createWithNumber({ views: { value: 1 } });

  const p = await actions.getWithNumberPermissionLiteral({ id: post.id });
  expect(p!.id).toEqual(post.id);
});

test("number permission on literal - not matching value - is not authorized", async () => {
  const post = await actions.createWithNumber({ views: { value: 100 } });

  await expect(
    actions.getWithNumberPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("number permission on literal - null value - is not authorized", async () => {
  const post = await actions.createWithNumber({ views: { isNull: true } });

  await expect(
    actions.getWithNumberPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("boolean permission on literal - matching value - is authorized", async () => {
  const post = await actions.createWithBoolean({ active: { value: true } });

  const p = await actions.getWithBooleanPermissionLiteral({ id: post.id });
  expect(p!.id).toEqual(post.id);
});

test("boolean permission on literal - not matching value - is not authorized", async () => {
  const post = await actions.createWithBoolean({ active: { value: false } });

  await expect(
    actions.getWithBooleanPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("boolean permission on literal - null value - is not authorized", async () => {
  const post = await actions.createWithBoolean({ active: { isNull: true } });

  await expect(
    actions.getWithBooleanPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("enum permission on literal - matching value - is authorized", async () => {
  const post = await actions.createWithEnum({
    type: { value: PostType.Technical },
  });

  const p = await actions.getWithEnumPermissionLiteral({ id: post.id });
  expect(p!.id).toEqual(post.id);
});

test("enum permission on literal - not matching value - is not authorized", async () => {
  const post = await actions.createWithEnum({
    type: { value: PostType.Lifestyle },
  });

  await expect(
    actions.getWithEnumPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("enum permission on literal - null value - is not authorized", async () => {
  const post = await actions.createWithEnum({ type: { isNull: true } });

  await expect(
    actions.getWithEnumPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("string permission on field - matching value - is authorized", async () => {
  const post = await actions.createWithText({ title: { value: "hello" } });

  const p = await actions.getWithTextPermissionFromField({ id: post.id });
  expect(p!.id).toEqual(post.id);
});

test("string permission on field - not matching value - is not authorized", async () => {
  const post = await actions.createWithText({ title: { value: "goodbye" } });

  await expect(
    actions.getWithTextPermissionFromField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("string permission on field - null value - is not authorized", async () => {
  const post = await actions.createWithText({ title: { isNull: true } });

  await expect(
    actions.getWithTextPermissionFromField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("number permission on field - matching value - is authorized", async () => {
  const post = await actions.createWithNumber({ views: { value: 1 } });

  const p = await actions.getWithNumberPermissionFromField({ id: post.id });
  expect(p!.id).toEqual(post.id);
});

test("number permission on field - not matching value - is not authorized", async () => {
  const post = await actions.createWithNumber({ views: { value: 100 } });

  await expect(
    actions.getWithNumberPermissionFromField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("number permission on field - null value - is not authorized", async () => {
  const post = await actions.createWithNumber({ views: { isNull: true } });

  await expect(
    actions.getWithNumberPermissionFromField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("boolean permission on field - matching value - is authorized", async () => {
  const post = await actions.createWithBoolean({ active: { value: true } });

  const p = await actions.getWithBooleanPermissionFromField({ id: post.id });
  expect(p!.id).toEqual(post.id);
});

test("boolean permission on field - unmatching value - field is not authorized", async () => {
  const post = await actions.createWithBoolean({ active: { value: false } });

  await expect(
    actions.getWithBooleanPermissionFromField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("boolean permission on field - null value - is not authorized", async () => {
  const post = await actions.createWithBoolean({ active: { isNull: true } });

  await expect(
    actions.getWithBooleanPermissionFromField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("enum permission on field name - matching value - is authorized", async () => {
  const post = await actions.createWithEnum({
    type: { value: PostType.Technical },
  });

  const p = await actions.getWithEnumPermissionFromField({ id: post.id });
  expect(p!.id).toEqual(post.id);
});

test("enum permission on field name - not matching value - is not authorized", async () => {
  const post = await actions.createWithEnum({
    type: { value: PostType.Lifestyle },
  });

  await expect(
    actions.getWithEnumPermissionFromField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("enum permission on field name - null value - is not authorized", async () => {
  const post = await actions.createWithEnum({ type: { isNull: true } });

  await expect(
    actions.getWithEnumPermissionFromField({ id: post.id })
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

  const p = await actions
    .withAuthToken(token)
    .getWithIdentityPermission({ id: post.id });
  expect(p!.id).toEqual(post.id);
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
    actions.withAuthToken(token2).getWithIdentityPermission({ id: post.id })
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
    actions.getWithIdentityPermission({ id: post.id })
  ).toHaveAuthorizationError();
});

test("true value permission - unauthenticated identity - is authorized", async () => {
  const post = await actions.createWithText({ title: { value: "hello" } });

  const p = await actions.getWithTrueValuePermission({ id: post.id });
  expect(p!.id).toEqual(post.id);
});
