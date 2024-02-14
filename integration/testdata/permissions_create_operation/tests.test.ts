import { test, expect, beforeEach } from "vitest";
import { actions, resetDatabase, models } from "@teamkeel/testing";
import { PostType } from "@teamkeel/sdk";

beforeEach(resetDatabase);

test("string permission on literal - matching value - is authorized", async () => {
  await expect(
    actions.createWithTextPermissionLiteral({ title: "hello" })
  ).resolves.toMatchObject({ title: "hello" });
});

test("string permission on literal - not matching value - is not authorized", async () => {
  await expect(
    actions.createWithTextPermissionLiteral({ title: "not hello" })
  ).toHaveAuthorizationError();
});

test("string permission on literal - null value - is not authorized", async () => {
  await expect(
    actions.createWithTextPermissionLiteral({ title: null })
  ).toHaveAuthorizationError();
});

test("number permission on literal - matching value - is authorized", async () => {
  await expect(
    actions.createWithNumberPermissioLiteral({ views: 5 })
  ).resolves.toMatchObject({ views: 5 });
});

test("number permission on literal - not matching value - is not authorized", async () => {
  await expect(
    actions.createWithNumberPermissioLiteral({ views: 500 })
  ).toHaveAuthorizationError();
});

test("number permission on literal - null value - is not authorized", async () => {
  await expect(
    actions.createWithNumberPermissioLiteral({ views: null })
  ).toHaveAuthorizationError();
});

test("boolean permission on literal - matching value - is authorized", async () => {
  await expect(
    actions.createWithBooleanPermissionLiteral({ active: true })
  ).resolves.toMatchObject({ active: true });
});

test("boolean permission on literal - not matching value - is not authorized", async () => {
  await expect(
    actions.createWithBooleanPermissionLiteral({ active: false })
  ).toHaveAuthorizationError();
});

test("boolean permission on literal - null value - is not authorized", async () => {
  await expect(
    actions.createWithBooleanPermissionLiteral({ active: null })
  ).toHaveAuthorizationError();
});

test("enum permission on literal - matching value - is authorized", async () => {
  await expect(
    actions.createWithEnumPermissionLiteral({ type: PostType.Technical })
  ).resolves.toMatchObject({ type: PostType.Technical });
});

test("enum permission on literal - not matching value - is not authorized", async () => {
  await expect(
    actions.createWithEnumPermissionLiteral({ type: PostType.Lifestyle })
  ).toHaveAuthorizationError();
});

test("enum permission on literal - null value - is not authorized", async () => {
  await expect(
    actions.createWithEnumPermissionLiteral({ type: null })
  ).toHaveAuthorizationError();
});

test("string permission on field name - matching value - is authorized", async () => {
  await expect(
    actions.createWithTextPermissionFromField({ title: "hello" })
  ).resolves.toMatchObject({ title: "hello" });
});

test("string permission on field name - not matching value - is not authorized", async () => {
  await expect(
    actions.createWithTextPermissionFromField({ title: "not hello" })
  ).toHaveAuthorizationError();
});

test("string permission on field name - null value - is not authorized", async () => {
  await expect(
    actions.createWithTextPermissionFromField({ title: null })
  ).toHaveAuthorizationError();
});

test("number permission on field name - matching value - is authorized", async () => {
  await expect(
    actions.createWithNumberPermissionFromField({ views: 5 })
  ).resolves.toMatchObject({ views: 5 });
});

test("number permission on field name - not matching value - is not authorized", async () => {
  await expect(
    actions.createWithNumberPermissionFromField({ views: 500 })
  ).toHaveAuthorizationError();
});

test("number permission on field name - null value - is not authorized", async () => {
  await expect(
    actions.createWithNumberPermissionFromField({ views: null })
  ).toHaveAuthorizationError();
});

test("boolean permission on field name - matching value - is authorized", async () => {
  await expect(
    actions.createWithBooleanPermissionFromField({ active: true })
  ).resolves.toMatchObject({ active: true });
});

test("boolean permission on field name - not matching value - is not authorized", async () => {
  await expect(
    actions.createWithBooleanPermissionFromField({ active: false })
  ).toHaveAuthorizationError();
});

test("boolean permission on field name - null value - is not authorized", async () => {
  await expect(
    actions.createWithBooleanPermissionFromField({ active: null })
  ).toHaveAuthorizationError();
});

test("enum permission on field name - matching value - is authorized", async () => {
  await expect(
    actions.createWithEnumPermissionFromField({ type: PostType.Technical })
  ).resolves.toMatchObject({ type: PostType.Technical });
});

test("enum permission on field name - not matching value - is not authorized", async () => {
  await expect(
    actions.createWithEnumPermissionFromField({ type: PostType.Lifestyle })
  ).toHaveAuthorizationError();
});

test("enum permission on field name - null value - is not authorized", async () => {
  await expect(
    actions.createWithEnumPermissionFromField({ type: null })
  ).toHaveAuthorizationError();
});

test("identity permission - correct identity in context - is authorized", async () => {
  const identity = await models.identity.create({
    email: "user@keel.xyz",
    password: "1234",
  });

  await expect(
    actions.withIdentity(identity).createWithIdentityRequiresSameIdentity({})
  ).resolves.toMatchObject({ id: expect.any(String) });
});

test("true value permission - with unauthenticated identity - is authorized", async () => {
  await expect(
    actions.createWithTrueValuePermission({ title: "hello" })
  ).resolves.toMatchObject({ title: "hello" });
});

test("multiple ORed permissions - matching a single value - is authorized", async () => {
  await expect(
    actions.createWithMultipleOrPermissions({
      title: "hello",
      views: 100,
      active: false,
    })
  ).resolves.toMatchObject({ title: "hello", views: 100, active: false });
});

test("multiple ORed permissions - matching all values - is authorized", async () => {
  await expect(
    actions.createWithMultipleOrPermissions({
      title: "hello",
      views: 5,
      active: true,
    })
  ).resolves.toMatchObject({ title: "hello", views: 5, active: true });
});

test("multiple ORed permissions - matching no values - is not authorized", async () => {
  await expect(
    actions.createWithMultipleOrPermissions({
      title: "goodbye",
      views: 100,
      active: false,
    })
  ).toHaveAuthorizationError();
});
