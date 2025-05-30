import { actions, models } from "@teamkeel/testing";
import { test, expect } from "vitest";
import { PostType } from "@teamkeel/sdk";

test("string permission on literal - matching value - is authorized", async () => {
  const post = await actions.createWithText({ title: "hello" });

  const p = await actions.getWithTextPermissionLiteral({ id: post.id });
  expect(p!.id).toEqual(post.id);
});

test("string permission on literal - not matching value - is not authorized", async () => {
  const post = await actions.createWithText({ title: "goodbye" });

  await expect(
    actions.getWithTextPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("string permission on literal - null value - is not authorized", async () => {
  const post = await actions.createWithText({ title: null });

  await expect(
    actions.getWithTextPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("number permission on literal - matching value - is authorized", async () => {
  const post = await actions.createWithNumber({ views: 1 });

  const p = await actions.getWithNumberPermissionLiteral({ id: post.id });
  expect(p!.id).toEqual(post.id);
});

test("number permission on literal - not matching value - is not authorized", async () => {
  const post = await actions.createWithNumber({ views: 100 });

  await expect(
    actions.getWithNumberPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("number permission on literal - null value - is not authorized", async () => {
  const post = await actions.createWithNumber({ views: null });

  await expect(
    actions.getWithNumberPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("boolean permission on literal - matching value - is authorized", async () => {
  const post = await actions.createWithBoolean({ active: true });

  const p = await actions.getWithBooleanPermissionLiteral({ id: post.id });
  expect(p!.id).toEqual(post.id);
});

test("boolean permission on literal - not matching value - is not authorized", async () => {
  const post = await actions.createWithBoolean({ active: false });

  await expect(
    actions.getWithBooleanPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("boolean permission on literal - null value - is not authorized", async () => {
  const post = await actions.createWithBoolean({ active: null });

  await expect(
    actions.getWithBooleanPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("enum permission on literal - matching value - is authorized", async () => {
  const post = await actions.createWithEnum({ type: PostType.Technical });

  const p = await actions.getWithEnumPermissionLiteral({ id: post.id });
  expect(p!.id).toEqual(post.id);
});

test("enum permission on literal - not matching value - is not authorized", async () => {
  const post = await actions.createWithEnum({ type: PostType.Lifestyle });

  await expect(
    actions.getWithEnumPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("enum permission on literal - null value - is not authorized", async () => {
  const post = await actions.createWithEnum({ type: null });

  await expect(
    actions.getWithEnumPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("string permission on field - matching value - is authorized", async () => {
  const post = await actions.createWithText({ title: "hello" });

  const p = await actions.getWithTextPermissionFromField({ id: post.id });
  expect(p!.id).toEqual(post.id);
});

test("string permission on field - not matching value - is not authorized", async () => {
  const post = await actions.createWithText({ title: "goodbye" });

  await expect(
    actions.getWithTextPermissionFromField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("string permission on field - null value - is not authorized", async () => {
  const post = await actions.createWithText({ title: null });

  await expect(
    actions.getWithTextPermissionFromField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("number permission on field - matching value - is authorized", async () => {
  const post = await actions.createWithNumber({ views: 1 });

  const p = await actions.getWithNumberPermissionFromField({ id: post.id });
  expect(p!.id).toEqual(post.id);
});

test("number permission on field - not matching value - is not authorized", async () => {
  const post = await actions.createWithNumber({ views: 100 });

  await expect(
    actions.getWithNumberPermissionFromField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("number permission on field - null value - is not authorized", async () => {
  const post = await actions.createWithNumber({ views: null });

  await expect(
    actions.getWithNumberPermissionFromField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("boolean permission on field - matching value - is authorized", async () => {
  const post = await actions.createWithBoolean({ active: true });

  const p = await actions.getWithBooleanPermissionFromField({ id: post.id });
  expect(p!.id).toEqual(post.id);
});

test("boolean permission on field - unmatching value - field is not authorized", async () => {
  const post = await actions.createWithBoolean({ active: false });

  await expect(
    actions.getWithBooleanPermissionFromField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("boolean permission on field - null value - is not authorized", async () => {
  const post = await actions.createWithBoolean({ active: null });

  await expect(
    actions.getWithBooleanPermissionFromField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("enum permission on field name - matching value - is authorized", async () => {
  const post = await actions.createWithEnum({ type: PostType.Technical });

  const p = await actions.getWithEnumPermissionFromField({ id: post.id });
  expect(p!.id).toEqual(post.id);
});

test("enum permission on field name - not matching value - is not authorized", async () => {
  const post = await actions.createWithEnum({ type: PostType.Lifestyle });

  await expect(
    actions.getWithEnumPermissionFromField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("enum permission on field name - null value - is not authorized", async () => {
  const post = await actions.createWithEnum({ type: null });

  await expect(
    actions.getWithEnumPermissionFromField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("identity permission - correct identity in context - is authorized", async () => {
  const identity = await models.identity.create({
    email: "user@keel.xyz",
    issuer: "https://keel.so",
  });

  const post = await actions.withIdentity(identity).createWithIdentity();

  const p = await actions
    .withIdentity(identity)
    .getWithIdentityPermission({ id: post.id });
  expect(p!.id).toEqual(post.id);
});

test("identity permission - incorrect identity in context - is not authorized", async () => {
  const identity1 = await models.identity.create({
    email: "user1@keel.xyz",
    issuer: "https://keel.so",
  });

  const identity2 = await models.identity.create({
    email: "user2@keel.xyz",
    issuer: "https://keel.so",
  });

  const post = await actions.withIdentity(identity1).createWithIdentity();

  await expect(
    actions.withIdentity(identity2).getWithIdentityPermission({ id: post.id })
  ).toHaveAuthorizationError();
});

test("identity permission - no identity in context - is not authorized", async () => {
  const identity = await models.identity.create({
    email: "user4@keel.xyz",
    issuer: "https://keel.so",
  });

  const post = await actions.withIdentity(identity).createWithIdentity();

  await expect(
    actions.getWithIdentityPermission({ id: post.id })
  ).toHaveAuthorizationError();
});

test("true value permission - unauthenticated identity - is authorized", async () => {
  const post = await actions.createWithText({ title: "hello" });

  const p = await actions.getWithTrueValuePermission({ id: post.id });
  expect(p!.id).toEqual(post.id);
});
