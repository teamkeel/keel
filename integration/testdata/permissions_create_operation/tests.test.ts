import { test, expect, beforeEach } from "vitest";
import { actions, resetDatabase, models } from "@teamkeel/testing";
import { PostType } from "@teamkeel/sdk";

beforeEach(resetDatabase);

test("string permission on literal - matching value - is authorized", async () => {
  await expect(
    actions.createWithTextPermissionLiteral({ title: { value: "hello" } })
  ).resolves.toMatchObject({ title: "hello" });
});

test("string permission on literal - not matching value - is not authorized", async () => {
  await expect(
    actions.createWithTextPermissionLiteral({ title: { value: "not hello" } })
  ).toHaveAuthorizationError();
});

test("string permission on literal - null value - is not authorized", async () => {
  await expect(
    actions.createWithTextPermissionLiteral({ title: { isNull: true } })
  ).toHaveAuthorizationError();
});

test("number permission on literal - matching value - is authorized", async () => {
  await expect(
    actions.createWithNumberPermissioLiteral({ views: { value: 5 } })
  ).resolves.toMatchObject({ views: 5 });
});

test("number permission on literal - not matching value - is not authorized", async () => {
  await expect(
    actions.createWithNumberPermissioLiteral({ views: { value: 500 } })
  ).toHaveAuthorizationError();
});

test("number permission on literal - null value - is not authorized", async () => {
  await expect(
    actions.createWithNumberPermissioLiteral({ views: { isNull: true } })
  ).toHaveAuthorizationError();
});

test("boolean permission on literal - matching value - is authorized", async () => {
  await expect(
    actions.createWithBooleanPermissionLiteral({ active: { value: true } })
  ).resolves.toMatchObject({ active: true });
});

test("boolean permission on literal - not matching value - is not authorized", async () => {
  await expect(
    actions.createWithBooleanPermissionLiteral({ active: { value: false } })
  ).toHaveAuthorizationError();
});

test("boolean permission on literal - null value - is not authorized", async () => {
  await expect(
    actions.createWithBooleanPermissionLiteral({ active: { isNull: true } })
  ).toHaveAuthorizationError();
});

test("enum permission on literal - matching value - is authorized", async () => {
  await expect(
    actions.createWithEnumPermissionLiteral({
      type: { value: PostType.Technical },
    })
  ).resolves.toMatchObject({ type: PostType.Technical });
});

test("enum permission on literal - not matching value - is not authorized", async () => {
  await expect(
    actions.createWithEnumPermissionLiteral({
      type: { value: PostType.Lifestyle },
    })
  ).toHaveAuthorizationError();
});

test("enum permission on literal - null value - is not authorized", async () => {
  await expect(
    actions.createWithEnumPermissionLiteral({ type: { isNull: true } })
  ).toHaveAuthorizationError();
});

test("string permission on field name - matching value - is authorized", async () => {
  await expect(
    actions.createWithTextPermissionFromField({ title: { value: "hello" } })
  ).resolves.toMatchObject({ title: "hello" });
});

test("string permission on field name - not matching value - is not authorized", async () => {
  await expect(
    actions.createWithTextPermissionFromField({ title: { value: "not hello" } })
  ).toHaveAuthorizationError();
});

test("string permission on field name - null value - is not authorized", async () => {
  await expect(
    actions.createWithTextPermissionFromField({ title: { isNull: true } })
  ).toHaveAuthorizationError();
});

test("number permission on field name - matching value - is authorized", async () => {
  await expect(
    actions.createWithNumberPermissionFromField({ views: { value: 5 } })
  ).resolves.toMatchObject({ views: 5 });
});

test("number permission on field name - not matching value - is not authorized", async () => {
  await expect(
    actions.createWithNumberPermissionFromField({ views: { value: 500 } })
  ).toHaveAuthorizationError();
});

test("number permission on field name - null value - is not authorized", async () => {
  await expect(
    actions.createWithNumberPermissionFromField({ views: { isNull: true } })
  ).toHaveAuthorizationError();
});

test("boolean permission on field name - matching value - is authorized", async () => {
  await expect(
    actions.createWithBooleanPermissionFromField({ active: { value: true } })
  ).resolves.toMatchObject({ active: true });
});

test("boolean permission on field name - not matching value - is not authorized", async () => {
  await expect(
    actions.createWithBooleanPermissionFromField({ active: { value: false } })
  ).toHaveAuthorizationError();
});

test("boolean permission on field name - null value - is not authorized", async () => {
  await expect(
    actions.createWithBooleanPermissionFromField({ active: { isNull: true } })
  ).toHaveAuthorizationError();
});

test("enum permission on field name - matching value - is authorized", async () => {
  await expect(
    actions.createWithEnumPermissionFromField({
      type: { value: PostType.Technical },
    })
  ).resolves.toMatchObject({ type: PostType.Technical });
});

test("enum permission on field name - not matching value - is not authorized", async () => {
  await expect(
    actions.createWithEnumPermissionFromField({
      type: { value: PostType.Lifestyle },
    })
  ).toHaveAuthorizationError();
});

test("enum permission on field name - null value - is not authorized", async () => {
  await expect(
    actions.createWithEnumPermissionFromField({ type: { isNull: true } })
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

  await expect(
    actions.withAuthToken(token).createWithIdentityRequiresSameIdentity({})
  ).resolves.toMatchObject({ id: expect.any(String) });
});

test("true value permission - with unauthenticated identity - is authorized", async () => {
  await expect(
    actions.createWithTrueValuePermission({ title: { value: "hello" } })
  ).resolves.toMatchObject({ title: "hello" });
});

test("multiple ORed permissions - matching a single value - is authorized", async () => {
  await expect(
    actions.createWithMultipleOrPermissions({
      title: { value: "hello" },
      views: { value: 100 },
      active: { value: false },
    })
  ).resolves.toMatchObject({ title: "hello", views: 100, active: false });
});

test("multiple ORed permissions - matching all values - is authorized", async () => {
  await expect(
    actions.createWithMultipleOrPermissions({
      title: { value: "hello" },
      views: { value: 5 },
      active: { value: true },
    })
  ).resolves.toMatchObject({ title: "hello", views: 5, active: true });
});

test("multiple ORed permissions - matching no values - is not authorized", async () => {
  await expect(
    actions.createWithMultipleOrPermissions({
      title: { value: "goodbye" },
      views: { value: 100 },
      active: { value: false },
    })
  ).toHaveAuthorizationError();
});
