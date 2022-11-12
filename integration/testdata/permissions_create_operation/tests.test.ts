import { test, expect, actions, Post, Identity } from "@teamkeel/testing";

test("string permission on literal - matching value - is authorized", async () => {
  expect(
    await actions.createWithTextPermissionLiteral({ title: "hello" })
  ).notToHaveAuthorizationError();
});

test("string permission on literal - not matching value - is not authorized", async () => {
  expect(
    await actions.createWithTextPermissionLiteral({ title: "not hello" })
  ).toHaveAuthorizationError();
});

test("string permission on literal - null value - is not authorized", async () => {
  expect(
    await actions.createWithTextPermissionLiteral({ title: null })
  ).toHaveAuthorizationError();
});

test("number permission on literal - matching value - is authorized", async () => {
  expect(
    await actions.createWithNumberPermissioLiteral({ views: 5 })
  ).notToHaveAuthorizationError();
});

test("number permission on literal - not matching value - is not authorized", async () => {
  expect(
    await actions.createWithNumberPermissioLiteral({ views: 500 })
  ).toHaveAuthorizationError();
});

test("number permission on literal - null value - is not authorized", async () => {
  expect(
    await actions.createWithNumberPermissioLiteral({ views: null })
  ).toHaveAuthorizationError();
});

test("boolean permission on literal - matching value - is authorized", async () => {
  expect(
    await actions.createWithBooleanPermissionLiteral({ active: true })
  ).notToHaveAuthorizationError();
});

test("boolean permission on literal - not matching value - is not authorized", async () => {
  expect(
    await actions.createWithBooleanPermissionLiteral({ active: false })
  ).toHaveAuthorizationError();
});

test("boolean permission on literal - null value - is not authorized", async () => {
  expect(
    await actions.createWithBooleanPermissionLiteral({ active: null })
  ).toHaveAuthorizationError();
});

test("enum permission on literal - matching value - is authorized", async () => {
  expect(
    await actions.createWithEnumPermissionLiteral({ type: "Technical" })
  ).notToHaveAuthorizationError();
});

test("enum permission on literal - not matching value - is not authorized", async () => {
  expect(
    await actions.createWithEnumPermissionLiteral({ type: "Lifestyle" })
  ).toHaveAuthorizationError();
});

test("enum permission on literal - null value - is not authorized", async () => {
  expect(
    await actions.createWithEnumPermissionLiteral({ type: null })
  ).toHaveAuthorizationError();
});

test("string permission on field name - matching value - is authorized", async () => {
  expect(
    await actions.createWithTextPermissionFromField({ title: "hello" })
  ).notToHaveAuthorizationError();
});

test("string permission on field name - not matching value - is not authorized", async () => {
  expect(
    await actions.createWithTextPermissionFromField({ title: "not hello" })
  ).toHaveAuthorizationError(true);
});

test("string permission on field name - null value - is not authorized", async () => {
  expect(
    await actions.createWithTextPermissionFromField({ title: null })
  ).toHaveAuthorizationError(true);
});

test("number permission on field name - matching value - is authorized", async () => {
  expect(
    await actions.createWithNumberPermissionFromField({ views: 5 })
  ).notToHaveAuthorizationError();
});

test("number permission on field name - not matching value - is not authorized", async () => {
  expect(
    await actions.createWithNumberPermissionFromField({ views: 500 })
  ).toHaveAuthorizationError();
});

test("number permission on field name - null value - is not authorized", async () => {
  expect(
    await actions.createWithNumberPermissionFromField({ views: null })
  ).toHaveAuthorizationError();
});

test("boolean permission on field name - matching value - is authorized", async () => {
  expect(
    await actions.createWithBooleanPermissionFromField({ active: true })
  ).notToHaveAuthorizationError();
});

test("boolean permission on field name - not matching value - is not authorized", async () => {
  expect(
    await actions.createWithBooleanPermissionFromField({ active: false })
  ).toHaveAuthorizationError();
});

test("boolean permission on field name - null value - is not authorized", async () => {
  expect(
    await actions.createWithBooleanPermissionFromField({ active: null })
  ).toHaveAuthorizationError();
});

test("enum permission on field name - matching value - is authorized", async () => {
  expect(
    await actions.createWithEnumPermissionFromField({ type: "Technical" })
  ).notToHaveAuthorizationError();
});

test("enum permission on field name - not matching value - is not authorized", async () => {
  expect(
    await actions.createWithEnumPermissionFromField({ type: "Lifestyle" })
  ).toHaveAuthorizationError();
});

test("enum permission on field name - null value - is not authorized", async () => {
  expect(
    await actions.createWithEnumPermissionFromField({ type: null })
  ).toHaveAuthorizationError();
});

test("identity permission - correct identity in context - is authorized", async () => {
  const { identityId } = await actions.authenticate({
    createIfNotExists: true,
    email: "user@keel.xyz",
    password: "1234",
  });

  const { object: identity } = await Identity.findOne({ id: identityId });

  expect(
    await actions
      .withIdentity(identity)
      .createWithIdentityRequiresSameIdentity({ title: "hello" })
  ).notToHaveAuthorizationError();
});

test("identity permission - missing identity in context - is not authorized", async () => {
  expect(
    await actions.createWithIdentityRequiresSameIdentity({ title: "hello" })
  ).notToHaveAuthorizationError();
});

test("true value permission - with unauthenticated identity - is authorized", async () => {
  expect(
    await actions.createWithTrueValuePermission({ title: "hello" })
  ).notToHaveAuthorizationError();
});

test("multiple ORed permissions - matching a single value - is authorized", async () => {
  expect(
    await actions.createWithMultipleOrPermissions({
      title: "hello",
      views: 100,
      active: false,
    })
  ).notToHaveAuthorizationError();
});

test("multiple ORed permissions - matching all values - is authorized", async () => {
  expect(
    await actions.createWithMultipleOrPermissions({
      title: "hello",
      views: 5,
      active: true,
    })
  ).notToHaveAuthorizationError();
});

test("multiple ORed permissions - matching no values - is not authorized", async () => {
  expect(
    await actions.createWithMultipleOrPermissions({
      title: "goodbye",
      views: 100,
      active: false,
    })
  ).toHaveAuthorizationError();
});

test("permission on implicit input - matching value - is authorized", async () => {
  expect(
    await actions.createWithPermissionFromImplicitInput({ title: "hello" })
  ).notToHaveAuthorizationError();
});

test("permission on implicit input - not matching value - is not authorized", async () => {
  expect(
    await actions.createWithPermissionFromImplicitInput({ title: "goodbye" })
  ).toHaveAuthorizationError();
});

test("permission on implicit input - null value - is not authorized", async () => {
  expect(
    await actions.createWithPermissionFromImplicitInput({ title: null })
  ).toHaveAuthorizationError();
});

test("permission on explicit input - matching value - is authorized", async () => {
  expect(
    await actions.createWithPermissionFromExplicitInput({ explTitle: "hello" })
  ).notToHaveAuthorizationError();
});

test("permission on explicit input - not matching value - is not authorized", async () => {
  expect(
    await actions.createWithPermissionFromExplicitInput({
      explTitle: "goodbye",
    })
  ).toHaveAuthorizationError();
});

test("permission on explicit input - null value - is not authorized", async () => {
  expect(
    await actions.createWithPermissionFromExplicitInput({ explTitle: null })
  ).toHaveAuthorizationError();
});
