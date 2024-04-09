import { test, expect, beforeEach } from "vitest";
import { actions, resetDatabase, models } from "@teamkeel/testing";
import { PostType } from "@teamkeel/sdk";

beforeEach(resetDatabase);

test("string permission on literal - all matching - is authorized", async () => {
  await actions.createWithText({ title: "hello" });
  await actions.createWithText({ title: "hello" });
  await actions.createWithText({ title: "hello" });
  await actions.createWithText({ title: "goodbye", isActive: false });
  await actions.createWithText({ title: null, isActive: false });

  const r = await actions.listWithTextPermissionLiteral({
    where: {
      isActive: { equals: true },
    },
  });
  expect(r.results.length).toEqual(3);
});

test("string permission on literal - one not matching value - is not authorized", async () => {
  await actions.createWithText({ title: "hello" });
  await actions.createWithText({ title: "goodbye" });
  await actions.createWithText({ title: "hello" });

  await expect(
    actions.listWithTextPermissionLiteral({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("string permission on literal - one not matching null value - is not authorized", async () => {
  await actions.createWithText({ title: "hello" });
  await actions.createWithText({ title: null });
  await actions.createWithText({ title: "hello" });

  await expect(
    actions.listWithTextPermissionLiteral({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("number permission on literal - all matching - is authorized", async () => {
  await actions.createWithNumber({ views: 1 });
  await actions.createWithNumber({ views: 1 });
  await actions.createWithNumber({ views: 1 });
  await actions.createWithNumber({ views: 100, isActive: false });
  await actions.createWithNumber({ views: null, isActive: false });

  const r = await actions.listWithNumberPermissionLiteral({
    where: {
      isActive: { equals: true },
    },
  });
  expect(r.results.length).toEqual(3);
});

test("number permission on literal - one not matching value - is not authorized", async () => {
  await actions.createWithNumber({ views: 1 });
  await actions.createWithNumber({ views: 100 });
  await actions.createWithNumber({ views: 1 });

  await expect(
    actions.listWithNumberPermissionLiteral({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("number permission on literal - one not matching null value - is not authorized", async () => {
  await actions.createWithNumber({ views: 1 });
  await actions.createWithNumber({ views: null });
  await actions.createWithNumber({ views: 1 });

  await expect(
    actions.listWithNumberPermissionLiteral({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("boolean permission on literal - all matching - is authorized", async () => {
  await actions.createWithBoolean({ active: true });
  await actions.createWithBoolean({ active: true });
  await actions.createWithBoolean({ active: true });
  await actions.createWithBoolean({ active: false, isActive: false });
  await actions.createWithBoolean({ active: null, isActive: false });

  const r = await actions.listWithBooleanPermissionLiteral({
    where: {
      isActive: { equals: true },
    },
  });
  expect(r.results.length).toEqual(3);
});

test("boolean permission on literal - one not matching value - is not authorized", async () => {
  await actions.createWithBoolean({ active: true });
  await actions.createWithBoolean({ active: false });
  await actions.createWithBoolean({ active: true });

  await expect(
    actions.listWithBooleanPermissionLiteral({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("boolean permission on literal - one not matching null value - is not authorized", async () => {
  await actions.createWithBoolean({ active: true });
  await actions.createWithBoolean({ active: null });
  await actions.createWithBoolean({ active: true });

  await expect(
    actions.listWithBooleanPermissionLiteral({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("enum permission on literal - all matching - is authorized", async () => {
  await actions.createWithEnum({ type: PostType.Technical });
  await actions.createWithEnum({ type: PostType.Technical });
  await actions.createWithEnum({ type: PostType.Technical });
  await actions.createWithEnum({ type: PostType.Food, isActive: false });
  await actions.createWithEnum({ type: null, isActive: false });

  const r = await actions.listWithEnumPermissionLiteral({
    where: {
      isActive: { equals: true },
    },
  });
  expect(r.results.length).toEqual(3);
});

test("enum permission on literal - one not matching value - is not authorized", async () => {
  await actions.createWithEnum({ type: PostType.Technical });
  await actions.createWithEnum({ type: PostType.Food });
  await actions.createWithEnum({ type: PostType.Technical });

  await expect(
    actions.listWithEnumPermissionLiteral({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("enum permission on literal - one not matching null value - is not authorized", async () => {
  await actions.createWithEnum({ type: PostType.Technical });
  await actions.createWithEnum({ type: null });
  await actions.createWithEnum({ type: PostType.Technical });

  await expect(
    actions.listWithEnumPermissionLiteral({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("string permission on field - all matching - is authorized", async () => {
  await actions.createWithText({ title: "hello" });
  await actions.createWithText({ title: "hello" });
  await actions.createWithText({ title: "hello" });
  await actions.createWithText({ title: "goodbye", isActive: false });
  await actions.createWithText({ title: null, isActive: false });

  const r = await actions.listWithTextPermissionFromField({
    where: {
      isActive: { equals: true },
    },
  });
  expect(r.results.length).toEqual(3);
});

test("string permission on field - one not matching value - is not authorized", async () => {
  await actions.createWithText({ title: "hello" });
  await actions.createWithText({ title: "goodbye" });
  await actions.createWithText({ title: "hello" });

  await expect(
    actions.listWithTextPermissionFromField({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("string permission on field - one not matching null value - is not authorized", async () => {
  await actions.createWithText({ title: "hello" });
  await actions.createWithText({ title: null });
  await actions.createWithText({ title: "hello" });

  await expect(
    actions.listWithTextPermissionFromField({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("number permission on field - all matching - is authorized", async () => {
  await actions.createWithNumber({ views: 1 });
  await actions.createWithNumber({ views: 1 });
  await actions.createWithNumber({ views: 1 });
  await actions.createWithNumber({ views: 100, isActive: false });
  await actions.createWithNumber({ views: null, isActive: false });

  const r = await actions.listWithNumberPermissionFromField({
    where: {
      isActive: { equals: true },
    },
  });
  expect(r.results.length).toEqual(3);
});

test("number permission on field - one not matching value - is not authorized", async () => {
  await actions.createWithNumber({ views: 1 });
  await actions.createWithNumber({ views: 100 });
  await actions.createWithNumber({ views: 1 });

  await expect(
    actions.listWithNumberPermissionFromField({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("number permission on field - one not matching null value - is not authorized", async () => {
  await actions.createWithNumber({ views: 1 });
  await actions.createWithNumber({ views: null });
  await actions.createWithNumber({ views: 1 });

  await expect(
    actions.listWithNumberPermissionFromField({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("boolean permission on field - all matching - is authorized", async () => {
  await actions.createWithBoolean({ active: true });
  await actions.createWithBoolean({ active: true });
  await actions.createWithBoolean({ active: true });
  await actions.createWithBoolean({ active: false, isActive: false });
  await actions.createWithBoolean({ active: null, isActive: false });

  const r = await actions.listWithBooleanPermissionFromField({
    where: {
      isActive: { equals: true },
    },
  });
  expect(r.results.length).toEqual(3);
});

test("boolean permission on field - one not matching value - field is not authorized", async () => {
  await actions.createWithBoolean({ active: true });
  await actions.createWithBoolean({ active: false });
  await actions.createWithBoolean({ active: true });

  await expect(
    actions.listWithBooleanPermissionFromField({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("boolean permission on field - one not matching null value - is not authorized", async () => {
  await actions.createWithBoolean({ active: true });
  await actions.createWithBoolean({ active: null });
  await actions.createWithBoolean({ active: true });

  await expect(
    actions.listWithBooleanPermissionFromField({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("enum permission on field name - all matching - is authorized", async () => {
  await actions.createWithEnum({ type: PostType.Technical });
  await actions.createWithEnum({ type: PostType.Technical });
  await actions.createWithEnum({ type: PostType.Technical });
  await actions.createWithEnum({ type: PostType.Food, isActive: false });
  await actions.createWithEnum({ type: null, isActive: false });

  await expect(
    actions.listWithEnumPermissionFromField({
      where: {
        isActive: { equals: true },
      },
    })
  ).not.toHaveAuthorizationError();
});

test("enum permission on field name - one not matching value - is not authorized", async () => {
  await actions.createWithEnum({ type: PostType.Technical });
  await actions.createWithEnum({ type: PostType.Food });
  await actions.createWithEnum({ type: PostType.Technical });

  await expect(
    actions.listWithEnumPermissionFromField({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("enum permission on field name - one not matching null value - is not authorized", async () => {
  await actions.createWithEnum({ type: PostType.Technical });
  await actions.createWithEnum({ type: null });
  await actions.createWithEnum({ type: PostType.Technical });

  await expect(
    actions.listWithEnumPermissionFromField({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("identity permission - correct identity in context - is authorized", async () => {
  const identity1 = await models.identity.create({
    email: "user@keel.xyz",
    issuer: "https://keel.so",
  });

  await actions.withIdentity(identity1).createWithIdentity({});

  await expect(
    actions
      .withIdentity(identity1)
      .listWithIdentityPermission({ where: { isActive: { equals: true } } })
  ).not.toHaveAuthorizationError();
});

test("identity permission - incorrect identity in context - is not authorized", async () => {
  const identity1 = await models.identity.create({
    email: "user1@keel.xyz",
    issuer: "https://keel.so",
  });

  const identity2 = await models.identity.create({
    email: "user@keel.xyz",
    issuer: "https://keel.so",
  });

  await actions.withIdentity(identity1).createWithIdentity({});
  await actions.withIdentity(identity2).createWithIdentity({});

  await expect(
    actions.withIdentity(identity2).listWithIdentityPermission({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("identity permission - no identity in context - is not authorized", async () => {
  const identity = await models.identity.create({
    email: "user@keel.xyz",
    issuer: "https://keel.so",
  });

  await actions.withIdentity(identity).createWithIdentity({});
  await actions.createWithIdentity({ isActive: false });

  await expect(
    actions.listWithIdentityPermission({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("true value permission - unauthenticated identity - is authorized", async () => {
  await actions.createWithText({ title: "hello" });

  await expect(
    actions.listWithTrueValuePermission({})
  ).not.toHaveAuthorizationError();
});
