import { test, expect, actions, Post, Identity } from "@teamkeel/testing";

test("string permission on literal - matching value - is authorized", async () => {
  const { object: post } = await actions.createWithText({ title: "hello" });

  expect(
    await actions.updateWithTextPermissionLiteral({
      where: { id: post.id },
      values: { title: "goodbye" },
    })
  ).notToHaveAuthorizationError();
});

test("string permission on literal - not matching value - is not authorized", async () => {
  const { object: post } = await actions.createWithText({ title: "goodbye" });

  expect(
    await actions.updateWithTextPermissionLiteral({
      where: { id: post.id },
      values: { title: "hello" },
    })
  ).toHaveAuthorizationError();
});

test("string permission on literal - null value - is not authorized", async () => {
  const { object: post } = await actions.createWithText({ title: "goodbye" });

  expect(
    await actions.updateWithTextPermissionLiteral({
      where: { id: post.id },
      values: { title: null },
    })
  ).toHaveAuthorizationError();
});

test("number permission on literal - matching value - is authorized", async () => {
  const { object: post } = await actions.createWithNumber({ views: 1 });

  expect(
    await actions.updateWithNumberPermissionLiteral({
      where: { id: post.id },
      values: { views: 100 },
    })
  ).notToHaveAuthorizationError();
});

test("number permission on literal - not matching value - is not authorized", async () => {
  const { object: post } = await actions.createWithNumber({ views: 100 });

  expect(
    await actions.updateWithNumberPermissionLiteral({
      where: { id: post.id },
      values: { views: 1 },
    })
  ).toHaveAuthorizationError();
});

test("number permission on literal - null value - is not authorized", async () => {
  const { object: post } = await actions.createWithNumber({ views: 100 });

  expect(
    await actions.updateWithNumberPermissionLiteral({
      where: { id: post.id },
      values: { views: null },
    })
  ).toHaveAuthorizationError();
});

test("boolean permission on literal - matching value - is authorized", async () => {
  const { object: post } = await actions.createWithBoolean({ active: true });

  expect(
    await actions.updateWithBooleanPermissionLiteral({
      where: { id: post.id },
      values: { active: false },
    })
  ).notToHaveAuthorizationError();
});

test("boolean permission on literal - not matching value - is not authorized", async () => {
  const { object: post } = await actions.createWithBoolean({ active: false });

  expect(
    await actions.updateWithBooleanPermissionLiteral({
      where: { id: post.id },
      values: { active: true },
    })
  ).toHaveAuthorizationError();
});

test("boolean permission on literal - null value - is not authorized", async () => {
  const { object: post } = await actions.createWithBoolean({ active: false });

  expect(
    await actions.updateWithBooleanPermissionLiteral({
      where: { id: post.id },
      values: { active: null },
    })
  ).toHaveAuthorizationError();
});

test("enum permission on literal - matching value - is authorized", async () => {
  const { object: post } = await actions.createWithEnum({ type: "Technical" });

  expect(
    await actions.updateWithEnumPermissionLiteral({
      where: { id: post.id },
      values: { title: "goodbye" },
    })
  ).notToHaveAuthorizationError();
});

test("enum permission on literal - not matching value - is not authorized", async () => {
  const { object: post } = await actions.createWithEnum({ type: "Lifestyle" });

  expect(
    await actions.updateWithEnumPermissionLiteral({
      where: { id: post.id },
      values: { title: "hello" },
    })
  ).toHaveAuthorizationError();
});

test("enum permission on literal - null value - is not authorized", async () => {
  const { object: post } = await actions.createWithEnum({ title: null });

  expect(
    await actions.updateWithEnumPermissionLiteral({
      where: { id: post.id },
      values: { title: null },
    })
  ).toHaveAuthorizationError();
});

test("string permission on field - matching value - is authorized", async () => {
  const { object: post } = await actions.createWithText({ title: "hello" });

  expect(
    await actions.updateWithTextPermissionFromField({
      where: { id: post.id },
      values: { title: "goodbye" },
    })
  ).notToHaveAuthorizationError();
});

test("string permission on field - not matching value - is not authorized", async () => {
  const { object: post } = await actions.createWithText({ title: "goodbye" });

  expect(
    await actions.updateWithTextPermissionFromField({
      where: { id: post.id },
      values: { title: "hello" },
    })
  ).toHaveAuthorizationError();
});

test("string permission on field - null value - is not authorized", async () => {
  const { object: post } = await actions.createWithText({ title: "goodbye" });

  expect(
    await actions.updateWithTextPermissionFromField({
      where: { id: post.id },
      values: { title: null },
    })
  ).toHaveAuthorizationError();
});

test("number permission on field - matching value - is authorized", async () => {
  const { object: post } = await actions.createWithNumber({ views: 1 });

  expect(
    await actions.updateWithNumberPermissionFromField({
      where: { id: post.id },
      values: { views: 100 },
    })
  ).notToHaveAuthorizationError();
});

test("number permission on field - not matching value - is not authorized", async () => {
  const { object: post } = await actions.createWithNumber({ views: 100 });

  expect(
    await actions.updateWithNumberPermissionFromField({
      where: { id: post.id },
      values: { views: 1 },
    })
  ).toHaveAuthorizationError();
});

test("number permission on field - null value - is not authorized", async () => {
  const { object: post } = await actions.createWithNumber({ views: 100 });

  expect(
    await actions.updateWithNumberPermissionFromField({
      where: { id: post.id },
      values: { views: null },
    })
  ).toHaveAuthorizationError();
});

test("boolean permission on field - matching value - is authorized", async () => {
  const { object: post } = await actions.createWithBoolean({ active: true });

  expect(
    await actions.updateWithBooleanPermissionFromField({
      where: { id: post.id },
      values: { active: false },
    })
  ).notToHaveAuthorizationError();
});

test("boolean permission on field - field is not authorized", async () => {
  const { object: post } = await actions.createWithBoolean({ active: false });

  expect(
    await actions.updateWithBooleanPermissionFromField({
      where: { id: post.id },
      values: { active: true },
    })
  ).toHaveAuthorizationError();
});

test("boolean permission on field - null - is not authorized", async () => {
  const { object: post } = await actions.createWithBoolean({ active: false });

  expect(
    await actions.updateWithBooleanPermissionFromField({
      where: { id: post.id },
      values: { active: null },
    })
  ).toHaveAuthorizationError();
});

test("enum permission on field - matching value - is authorized", async () => {
  const { object: post } = await actions.createWithEnum({ type: "Technical" });

  expect(
    await actions.updateWithEnumPermissionFromField({
      where: { id: post.id },
      values: { type: "Lifestyle" },
    })
  ).notToHaveAuthorizationError();
});

test("enum permission on field - field is not authorized", async () => {
  const { object: post } = await actions.createWithEnum({ type: "Lifestyle" });

  expect(
    await actions.updateWithEnumPermissionFromField({
      where: { id: post.id },
      values: { type: "Technical" },
    })
  ).toHaveAuthorizationError();
});

test("enum permission on field - null - is not authorized", async () => {
  const { object: post } = await actions.createWithEnum({ type: null });

  expect(
    await actions.updateWithEnumPermissionFromField({
      where: { id: post.id },
      values: { type: null },
    })
  ).toHaveAuthorizationError();
});

test("identity permission - correct identity in context - is authorized", async () => {
  const { identityId } = await actions.authenticate({
    createIfNotExists: true,
    email: "user@keel.xyz",
    password: "1234",
  });

  const { object: identity } = await Identity.findOne({ id: identityId });

  const { object: post } = await actions
    .withIdentity(identity)
    .createWithIdentity({});

  expect(
    await actions.withIdentity(identity).updateWithIdentityPermission({
      where: { id: post.id },
      values: { title: "hello" },
    })
  ).notToHaveAuthorizationError();
});

test("identity permission - incorrect identity in context - is not authorized", async () => {
  const { identityId: id1 } = await actions.authenticate({
    createIfNotExists: true,
    email: "user1@keel.xyz",
    password: "1234",
  });

  const { identityId: id2 } = await actions.authenticate({
    createIfNotExists: true,
    email: "user2@keel.xyz",
    password: "1234",
  });

  const { object: identity1 } = await Identity.findOne({ id: id1 });
  const { object: identity2 } = await Identity.findOne({ id: id2 });

  const { object: post } = await actions
    .withIdentity(identity1)
    .createWithIdentity({});

  expect(
    await actions.withIdentity(identity2).updateWithIdentityPermission({
      where: { id: post.id },
      values: { title: "hello" },
    })
  ).toHaveAuthorizationError();
});

test("identity permission - no identity in context - is not authorized", async () => {
  const { identityId: id } = await actions.authenticate({
    createIfNotExists: true,
    email: "user@keel.xyz",
    password: "1234",
  });

  const { object: identity } = await Identity.findOne({ id: id });

  const { object: post } = await actions
    .withIdentity(identity)
    .createWithIdentity({});

  expect(
    await actions.updateWithIdentityPermission({
      where: { id: post.id },
      values: { title: "hello" },
    })
  ).toHaveAuthorizationError();
});

test("true value permission - unauthenticated identity - is authorized", async () => {
  const { object: post } = await actions.createWithText({ title: "hello" });

  expect(
    await actions.updateWithTrueValuePermission({
      where: { id: post.id },
      values: { title: "hello again" },
    })
  ).notToHaveAuthorizationError();
});

test("permission on implicit input - matching value - is authorized", async () => {
  const { object: post } = await actions.createWithText({ title: "hello" });

  expect(
    await actions.updateWithTextPermissionOnImplicitInput({
      where: { id: post.id },
      values: { title: "does not matter" },
    })
  ).notToHaveAuthorizationError();
});

test("permission on implicit input - not matching value - is not authorized", async () => {
  const { object: post } = await actions.createWithText({ title: "goodbye" });

  expect(
    await actions.updateWithTextPermissionOnImplicitInput({
      where: { id: post.id },
      values: { title: "does not matter" },
    })
  ).toHaveAuthorizationError();
});

test("permission on implicit input - null value - is not authorized", async () => {
  const { object: post } = await actions.createWithText({ title: null });

  expect(
    await actions.updateWithTextPermissionOnImplicitInput({
      where: { id: post.id },
      values: { title: "does not matter" },
    })
  ).toHaveAuthorizationError();
});

test("permission on explicit input - matching value - is authorized", async () => {
  const { object: post } = await actions.createWithText({ title: "hello" });

  expect(
    await actions.updateWithTextPermissionOnExplicitInput({
      where: { id: post.id },
      values: { explTitle: "hello" },
    })
  ).notToHaveAuthorizationError();
});

test("permission on explicit input - not matching value - is not authorized", async () => {
  const { object: post } = await actions.createWithText({ title: "hello" });

  expect(
    await actions.updateWithTextPermissionOnExplicitInput({
      where: { id: post.id },
      values: { explTitle: "goodbye" },
    })
  ).toHaveAuthorizationError();
});

test("permission on explicit input - null value - is not authorized", async () => {
  const { object: post } = await actions.createWithText({ title: "hello" });

  expect(
    await actions.updateWithTextPermissionOnExplicitInput({
      where: { id: post.id },
      values: { explTitle: null },
    })
  ).toHaveAuthorizationError();
});

test("permission on explicit input - matching null value - is not authorized", async () => {
  const { object: post } = await actions.createWithText({ title: null });

  expect(
    await actions.updateWithTextPermissionOnExplicitInput({
      where: { id: post.id },
      values: { explTitle: null },
    })
  ).notToHaveAuthorizationError();
});
