import { test, expect, actions, Identity } from "@teamkeel/testing";

test("string permission on literal - matching value - is authorized", async () => {
  const { object: post } = await actions.createWithTitle({ title: "hello" });

  expect(
    await actions.deleteWithTextPermissionLiteral({ id: post.id })
  ).notToHaveAuthorizationError();
});

test("string permission on literal - not matching value - is not authorized", async () => {
  const { object: post } = await actions.createWithTitle({
    title: "not hello",
  });

  expect(
    await actions.deleteWithTextPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("string permission on literal - not matching null value - is not authorized", async () => {
  const { object: post } = await actions.createWithTitle({ title: null });

  expect(
    await actions.deleteWithTextPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("number permission on literal - matching value - is authorized", async () => {
  const { object: post } = await actions.createWithViews({ views: 5 });

  expect(
    await actions.deleteWithNumberPermissionLiteral({ id: post.id })
  ).notToHaveAuthorizationError();
});

test("number permission on literal - not matching value - is not authorized", async () => {
  const { object: post } = await actions.createWithViews({ views: 500 });

  expect(
    await actions.deleteWithNumberPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("number permission on literal - not matching null value - is not authorized", async () => {
  const { object: post } = await actions.createWithViews({ views: null });

  expect(
    await actions.deleteWithNumberPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("boolean permission on literal - matching value - is authorized", async () => {
  const { object: post } = await actions.createWithActive({ active: true });

  expect(
    await actions.deleteWithBooleanPermissionLiteral({ id: post.id })
  ).notToHaveAuthorizationError();
});

test("boolean permission on literal - not matching value - is not authorized", async () => {
  const { object: post } = await actions.createWithActive({ active: false });

  expect(
    await actions.deleteWithBooleanPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("boolean permission on literal - not matching null value - is not authorized", async () => {
  const { object: post } = await actions.createWithActive({ active: null });

  expect(
    await actions.deleteWithBooleanPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("enum permission on literal - matching value - is authorized", async () => {
  const { object: post } = await actions.createWithEnum({ type: "Technical" });

  expect(
    await actions.deleteWithEnumPermissionLiteral({ id: post.id })
  ).notToHaveAuthorizationError();
});

test("enum permission on literal - not matching value - is not authorized", async () => {
  const { object: post } = await actions.createWithEnum({ title: "Lifestyle" });

  expect(
    await actions.deleteWithEnumPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("enum permission on literal - not matching null value - is not authorized", async () => {
  const { object: post } = await actions.createWithEnum({ title: null });

  expect(
    await actions.deleteWithEnumPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();
});

test("string permission on field - matching value - is authorized", async () => {
  const { object: post } = await actions.createWithTitle({ title: "hello" });

  expect(
    await actions.deleteWithTextPermissionOnField({ id: post.id })
  ).notToHaveAuthorizationError();
});

test("string permission on field - not matching value - is not authorized", async () => {
  const { object: post } = await actions.createWithTitle({
    title: "not hello",
  });

  expect(
    await actions.deleteWithTextPermissionOnField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("string permission on field - not matching null value - is not authorized", async () => {
  const { object: post } = await actions.createWithTitle({ title: null });

  expect(
    await actions.deleteWithTextPermissionOnField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("number permission on field - matching value - is authorized", async () => {
  const { object: post } = await actions.createWithViews({ views: 5 });

  expect(
    await actions.deleteWithNumberPermissionOnField({ id: post.id })
  ).notToHaveAuthorizationError();
});

test("number permission on field - not matching value - is not authorized", async () => {
  const { object: post } = await actions.createWithViews({ views: 500 });

  expect(
    await actions.deleteWithNumberPermissionOnField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("number permission on field - not matching null value - is not authorized", async () => {
  const { object: post } = await actions.createWithViews({ views: null });

  expect(
    await actions.deleteWithNumberPermissionOnField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("boolean permission on field - matching value - is authorized", async () => {
  const { object: post } = await actions.createWithActive({ active: true });

  expect(
    await actions.deleteWithBooleanPermissionOnField({ id: post.id })
  ).notToHaveAuthorizationError();
});

test("boolean permission on field - not matching value - is not authorized", async () => {
  const { object: post } = await actions.createWithActive({ active: false });

  expect(
    await actions.deleteWithBooleanPermissionOnField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("boolean permission on field - not matching null value - is not authorized", async () => {
  const { object: post } = await actions.createWithActive({ active: null });

  expect(
    await actions.deleteWithBooleanPermissionOnField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("enum permission on field - matching value - is authorized", async () => {
  const { object: post } = await actions.createWithEnum({ type: "Technical" });

  expect(
    await actions.deleteWithEnumPermissionOnField({ id: post.id })
  ).notToHaveAuthorizationError();
});

test("enum permission on field - not matching value - is not authorized", async () => {
  const { object: post } = await actions.createWithEnum({ type: "Lifestyle" });

  expect(
    await actions.deleteWithEnumPermissionOnField({ id: post.id })
  ).toHaveAuthorizationError();
});

test("enum permission on field - not matching null value - is not authorized", async () => {
  const { object: post } = await actions.createWithEnum({ type: null });

  expect(
    await actions.deleteWithEnumPermissionOnField({ id: post.id })
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
      .deleteWithRequiresSameIdentity({ id: post.id })
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
      .deleteWithRequiresSameIdentity({ id: post.id })
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
    await actions.deleteWithRequiresSameIdentity({ id: post.id })
  ).toHaveAuthorizationError();
});

test("true value permission - with unauthenticated identity - is authorized", async () => {
  const { object: post } = await actions.createWithActive({ active: true });

  expect(
    await actions.deleteWithTrueValuePermission({ id: post.id })
  ).notToHaveAuthorizationError();
});

test("permission on implicit input - matching value - is authorized", async () => {
  const { object: post } = await actions.createWithTitle({ title: "hello" });

  expect(
    await actions.deleteWithPermissionOnImplicitInput({
      id: post.id,
      title: "hello",
    })
  ).notToHaveAuthorizationError();
});

test("permission on implicit input - not matching value - is not authorized", async () => {
  const { object: post } = await actions.createWithTitle({ title: "hello" });

  expect(
    await actions.deleteWithPermissionOnImplicitInput({
      id: post.id,
      title: "goodbye",
    })
  ).toHaveAuthorizationError();
});

test("permission on implicit input - not matching null value - is not authorized", async () => {
  const { object: post } = await actions.createWithTitle({ title: "hello" });

  expect(
    await actions.deleteWithPermissionOnImplicitInput({
      id: post.id,
      title: null,
    })
  ).toHaveAuthorizationError();
});

test("permission on explicit input - not matching value - is not authorized", async () => {
  const { object: post } = await actions.createWithTitle({ title: "hello" });

  expect(
    await actions.deleteWithPermissionOnExplicitInput({
      id: post.id,
      explTitle: "goodbye",
    })
  ).toHaveAuthorizationError();
});

test("permission on explicit input - matching null value - is authorized", async () => {
  const { object: post } = await actions.createWithTitle({ title: null });

  expect(
    await actions.deleteWithPermissionOnExplicitInput({
      id: post.id,
      explTitle: null,
    })
  ).notToHaveAuthorizationError();
});

test("permission on explicit input - not matching null value - is not authorized", async () => {
  const { object: post } = await actions.createWithTitle({ title: "hello" });

  expect(
    await actions.deleteWithPermissionOnExplicitInput({
      id: post.id,
      explTitle: null,
    })
  ).toHaveAuthorizationError();
});
