import { test, expect, actions, Post, Identity } from "@teamkeel/testing";

test("string permission on literal - matching value - is authorized", async () => {
  const { object: post } = await actions.createWithText({ title: "hello" });

  expect(
    await actions.getWithTextPermissionLiteral({ id: post.id })
  ).notToHaveAuthorizationError();
});

test("string permission on literal - not matching value - is not authorized", async () => {
  const { object: post } = await actions.createWithText({ title: "goodbye" });

  expect(
    await actions.getWithTextPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("string permission on literal - null value - is not authorized", async () => {
  const { object: post } = await actions.createWithText({ title: null });

  expect(
    await actions.getWithTextPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("number permission on literal - matching value - is authorized", async () => {
  const { object: post } = await actions.createWithNumber({ views: 1 });

  expect(
    await actions.getWithNumberPermissionLiteral({ id: post.id })
  ).notToHaveAuthorizationError();
});

test("number permission on literal - not matching value - is not authorized", async () => {
  const { object: post } = await actions.createWithNumber({ views: 100 });

  expect(
    await actions.getWithNumberPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("number permission on literal - null value - is not authorized", async () => {
  const { object: post } = await actions.createWithNumber({ views: null });

  expect(
    await actions.getWithNumberPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("boolean permission on literal - matching value - is authorized", async () => {
  const { object: post } = await actions.createWithBoolean({ active: true });

  expect(
    await actions.getWithBooleanPermissionLiteral({ id: post.id })
  ).notToHaveAuthorizationError();
});

test("boolean permission on literal - not matching value - is not authorized", async () => {
  const { object: post } = await actions.createWithBoolean({ active: false });

  expect(
    await actions.getWithBooleanPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("boolean permission on literal - null value - is not authorized", async () => {
  const { object: post } = await actions.createWithBoolean({ active: null });

  expect(
    await actions.getWithBooleanPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("enum permission on literal - matching value - is authorized", async () => {
  const { object: post } = await actions.createWithEnum({ type: "Technical" });

  expect(
    await actions.getWithEnumPermissionLiteral({ id: post.id })
  ).notToHaveAuthorizationError();
});

test("enum permission on literal - not matching value - is not authorized", async () => {
  const { object: post } = await actions.createWithEnum({ type: "Lifestyle" });

  expect(
    await actions.getWithEnumPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("enum permission on literal - null value - is not authorized", async () => {
  const { object: post } = await actions.createWithEnum({ title: null });

  expect(
    await actions.getWithEnumPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("string permission on field - matching value - is authorized", async () => {
  const { object: post } = await actions.createWithText({ title: "hello" });

  expect(
    await actions.getWithTextPermissionFromField({ id: post.id })
  ).notToHaveAuthorizationError();
});

test("string permission on field - not matching value - is not authorized", async () => {
  const { object: post } = await actions.createWithText({ title: "goodbye" });

  expect(
    await actions.getWithTextPermissionFromField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("string permission on field - null value - is not authorized", async () => {
  const { object: post } = await actions.createWithText({ title: null });

  expect(
    await actions.getWithTextPermissionFromField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("number permission on field - matching value - is authorized", async () => {
  const { object: post } = await actions.createWithNumber({ views: 1 });

  expect(
    await actions.getWithNumberPermissionFromField({ id: post.id })
  ).notToHaveAuthorizationError();
});

test("number permission on field - not matching value - is not authorized", async () => {
  const { object: post } = await actions.createWithNumber({ views: 100 });

  expect(
    await actions.getWithNumberPermissionFromField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("number permission on field - null value - is not authorized", async () => {
  const { object: post } = await actions.createWithNumber({ views: null });

  expect(
    await actions.getWithNumberPermissionFromField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("boolean permission on field - matching value - is authorized", async () => {
  const { object: post } = await actions.createWithBoolean({ active: true });

  expect(
    await actions.getWithBooleanPermissionFromField({ id: post.id })
  ).notToHaveAuthorizationError();
});

test("boolean permission on field - unmatching value - field is not authorized", async () => {
  const { object: post } = await actions.createWithBoolean({ active: false });

  expect(
    await actions.getWithBooleanPermissionFromField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("boolean permission on field - null value - is not authorized", async () => {
  const { object: post } = await actions.createWithBoolean({ active: null });

  expect(
    await actions.getWithBooleanPermissionFromField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("enum permission on field name - matching value - is authorized", async () => {
  const { object: post } = await actions.createWithEnum({ type: "Technical" });

  expect(
    await actions.getWithEnumPermissionFromField({ id: post.id })
  ).notToHaveAuthorizationError();
});

test("enum permission on field name - not matching value - is not authorized", async () => {
  const { object: post } = await actions.createWithEnum({ type: "Lifestyle" });

  expect(
    await actions.getWithEnumPermissionFromField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("enum permission on field name - null value - is not authorized", async () => {
  const { object: post } = await actions.createWithEnum({ type: null });

  expect(
    await actions.getWithEnumPermissionFromField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("identity permission - correct identity in context - is authorized", async () => {
  const { identityId } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "user@keel.xyz",
      password: "1234",
    },
  });

  const { object: identity } = await Identity.findOne({ id: identityId });

  const { object: post } = await actions
    .withIdentity(identity)
    .createWithIdentity({});

  expect(
    await actions
      .withIdentity(identity)
      .getWithIdentityPermission({ id: post.id })
  ).notToHaveAuthorizationError();
});

test("identity permission - incorrect identity in context - is not authorized", async () => {
  const { identityId: id1 } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "user1@keel.xyz",
      password: "1234",
    },
  });

  const { identityId: id2 } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "user2@keel.xyz",
      password: "1234",
    },
  });

  const { object: identity1 } = await Identity.findOne({ id: id1 });
  const { object: identity2 } = await Identity.findOne({ id: id2 });

  const { object: post } = await actions
    .withIdentity(identity1)
    .createWithIdentity({});

  expect(
    await actions
      .withIdentity(identity2)
      .getWithIdentityPermission({ id: post.id })
  ).toHaveAuthorizationError();
});

test("identity permission - no identity in context - is not authorized", async () => {
  const { identityId: id } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "user@keel.xyz",
      password: "1234",
    },
  });

  const { object: identity } = await Identity.findOne({ id: id });

  const { object: post } = await actions
    .withIdentity(identity)
    .createWithIdentity({});

  expect(
    await actions.getWithIdentityPermission({ id: post.id })
  ).toHaveAuthorizationError();
});

test("true value permission - unauthenticated identity - is authorized", async () => {
  const { object: post } = await actions.createWithText({ title: "hello" });

  expect(
    await actions.getWithTrueValuePermission({ id: post.id })
  ).notToHaveAuthorizationError();
});

test("permission on implicit input - matching value - is authorized", async () => {
  const { object: post } = await actions.createWithText({ title: "hello" });

  expect(
    await actions.getWithTextPermissionFromImplicitInput({ id: post.id })
  ).notToHaveAuthorizationError();
});

test("permission on implicit input - not matching value - is not authorized", async () => {
  const { object: post } = await actions.createWithText({ title: "hello" });

  expect(
    await actions.getWithTextPermissionFromImplicitInputNotMatching({
      id: post.id,
    })
  ).toHaveAuthorizationError();
});

test("permission on explicit input - matching value - is authorized", async () => {
  const { object: post } = await actions.createWithText({ title: "hello" });

  expect(
    await actions.getWithTextPermissionFromExplicitInput({
      id: post.id,
      explTitle: "hello",
    })
  ).notToHaveAuthorizationError();
});

test("permission on explicit input - not matching value - is not authorized", async () => {
  const { object: post } = await actions.createWithText({ title: "hello" });

  expect(
    await actions.getWithTextPermissionFromExplicitInput({
      id: post.id,
      explTitle: "goodbye",
    })
  ).toHaveAuthorizationError();
});

test("permission on explicit input - not matching null value - is not authorized", async () => {
  const { object: post } = await actions.createWithText({ title: "hello" });

  expect(
    await actions.getWithTextPermissionFromExplicitInput({
      id: post.id,
      explTitle: null,
    })
  ).toHaveAuthorizationError();
});

test("permission on explicit input - matching null value - is not authorized", async () => {
  const { object: post } = await actions.createWithText({ title: null });

  expect(
    await actions.getWithTextPermissionFromExplicitInput({
      id: post.id,
      explTitle: null,
    })
  ).notToHaveAuthorizationError();
});
