import { test, expect, actions, Post, Identity } from "@teamkeel/testing";

test("string permission on literal - all matching - is authorized", async () => {
  await actions.createWithText({ title: "hello" });
  await actions.createWithText({ title: "hello" });
  await actions.createWithText({ title: "hello" });
  await actions.createWithText({ title: "goodbye", isActive: false });
  await actions.createWithText({ title: null, isActive: false });

  expect(
    await actions.listWithTextPermissionLiteral({
      where: {
        isActive: { equals: true },
      },
    })
  ).notToHaveAuthorizationError();
});

test("string permission on literal - one not matching value - is not authorized", async () => {
  await actions.createWithText({ title: "hello" });
  await actions.createWithText({ title: "goodbye" });
  await actions.createWithText({ title: "hello" });

  expect(
    await actions.listWithTextPermissionLiteral({
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

  expect(
    await actions.listWithTextPermissionLiteral({
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

  expect(
    await actions.listWithNumberPermissionLiteral({
      where: {
        isActive: { equals: true },
      },
    })
  ).notToHaveAuthorizationError();
});

test("number permission on literal - one not matching value - is not authorized", async () => {
  await actions.createWithNumber({ views: 1 });
  await actions.createWithNumber({ views: 100 });
  await actions.createWithNumber({ views: 1 });

  expect(
    await actions.listWithNumberPermissionLiteral({
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

  expect(
    await actions.listWithNumberPermissionLiteral({
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

  expect(
    await actions.listWithBooleanPermissionLiteral({
      where: {
        isActive: { equals: true },
      },
    })
  ).notToHaveAuthorizationError();
});

test("boolean permission on literal - one not matching value - is not authorized", async () => {
  await actions.createWithBoolean({ active: true });
  await actions.createWithBoolean({ active: false });
  await actions.createWithBoolean({ active: true });

  expect(
    await actions.listWithBooleanPermissionLiteral({
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

  expect(
    await actions.listWithBooleanPermissionLiteral({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("enum permission on literal - all matching - is authorized", async () => {
  await actions.createWithEnum({ type: "Technical" });
  await actions.createWithEnum({ type: "Technical" });
  await actions.createWithEnum({ type: "Technical" });
  await actions.createWithEnum({ type: "Food", isActive: false });
  await actions.createWithEnum({ type: null, isActive: false });

  expect(
    await actions.listWithEnumPermissionLiteral({
      where: {
        isActive: { equals: true },
      },
    })
  ).notToHaveAuthorizationError();
});

test("enum permission on literal - one not matching value - is not authorized", async () => {
  await actions.createWithEnum({ type: "Technical" });
  await actions.createWithEnum({ type: "Food" });
  await actions.createWithEnum({ type: "Technical" });

  expect(
    await actions.listWithEnumPermissionLiteral({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("enum permission on literal - one not matching null value - is not authorized", async () => {
  await actions.createWithEnum({ type: "Technical" });
  await actions.createWithEnum({ type: null });
  await actions.createWithEnum({ type: "Technical" });

  expect(
    await actions.listWithEnumPermissionLiteral({
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

  expect(
    await actions.listWithTextPermissionFromField({
      where: {
        isActive: { equals: true },
      },
    })
  ).notToHaveAuthorizationError();
});

test("string permission on field - one not matching value - is not authorized", async () => {
  await actions.createWithText({ title: "hello" });
  await actions.createWithText({ title: "goodbye" });
  await actions.createWithText({ title: "hello" });

  expect(
    await actions.listWithTextPermissionFromField({
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

  expect(
    await actions.listWithTextPermissionFromField({
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

  expect(
    await actions.listWithNumberPermissionFromField({
      where: {
        isActive: { equals: true },
      },
    })
  ).notToHaveAuthorizationError();
});

test("number permission on field - one not matching value - is not authorized", async () => {
  await actions.createWithNumber({ views: 1 });
  await actions.createWithNumber({ views: 100 });
  await actions.createWithNumber({ views: 1 });

  expect(
    await actions.listWithNumberPermissionFromField({
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

  expect(
    await actions.listWithNumberPermissionFromField({
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

  expect(
    await actions.listWithBooleanPermissionFromField({
      where: {
        isActive: { equals: true },
      },
    })
  ).notToHaveAuthorizationError();
});

test("boolean permission on field - one not matching value - field is not authorized", async () => {
  await actions.createWithBoolean({ active: true });
  await actions.createWithBoolean({ active: false });
  await actions.createWithBoolean({ active: true });

  expect(
    await actions.listWithBooleanPermissionFromField({
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

  expect(
    await actions.listWithBooleanPermissionFromField({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("enum permission on field name - all matching - is authorized", async () => {
  await actions.createWithEnum({ type: "Technical" });
  await actions.createWithEnum({ type: "Technical" });
  await actions.createWithEnum({ type: "Technical" });
  await actions.createWithEnum({ type: "Food", isActive: false });
  await actions.createWithEnum({ type: null, isActive: false });

  expect(
    await actions.listWithEnumPermissionFromField({
      where: {
        isActive: { equals: true },
      },
    })
  ).notToHaveAuthorizationError();
});

test("enum permission on field name - one not matching value - is not authorized", async () => {
  await actions.createWithEnum({ type: "Technical" });
  await actions.createWithEnum({ type: "Food" });
  await actions.createWithEnum({ type: "Technical" });

  expect(
    await actions.listWithEnumPermissionFromField({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("enum permission on field name - one not matching null value - is not authorized", async () => {
  await actions.createWithEnum({ type: "Technical" });
  await actions.createWithEnum({ type: null });
  await actions.createWithEnum({ type: "Technical" });

  expect(
    await actions.listWithEnumPermissionFromField({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("identity permission - correct identity in context - is authorized", async () => {
  const { identityId: id1 } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "user@keel.xyz",
      password: "1234",
    },
  });

  const { object: identity } = await Identity.findOne({ id: id1 });

  await actions.withIdentity(identity).createWithIdentity({});

  await actions.withIdentity(identity).createWithIdentity({});

  const { identityId: id2 } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "anotheruser@keel.xyz",
      password: "beep",
    },
  });

  const { object: identity2 } = await Identity.findOne({ id: id2 });

  await actions.withIdentity(identity2).createWithIdentity({ isActive: false });

  expect(
    await actions
      .withIdentity(identity)
      .listWithIdentityPermission({ isActive: { equals: true } })
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

  await actions.withIdentity(identity1).createWithIdentity({});

  await actions.withIdentity(identity2).createWithIdentity({});

  expect(
    await actions.withIdentity(identity2).listWithIdentityPermission({
      where: {
        isActive: { equals: true },
      },
    })
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

  await actions.withIdentity(identity).createWithIdentity({});

  await actions.createWithIdentity({ isActive: false });

  expect(
    await actions.listWithIdentityPermission({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("true value permission - unauthenticated identity - is authorized", async () => {
  await actions.createWithText({ title: "hello" });

  expect(
    await actions.listWithTrueValuePermission({})
  ).notToHaveAuthorizationError();
});

test("permission on explicit input - all matching - is authorized", async () => {
  await actions.createWithText({ title: "hello" });
  await actions.createWithText({ title: "hello" });
  await actions.createWithText({ title: "hello" });
  await actions.createWithText({ title: "goodbye", isActive: false });
  await actions.createWithText({ title: null, isActive: false });

  expect(
    await actions.listWithTextPermissionFromExplicitInput({
      where: {
        isActive: { equals: true },
        explTitle: "hello",
      },
    })
  ).notToHaveAuthorizationError();
});

test("permission on explicit input - one not matching - is not authorized", async () => {
  await actions.createWithText({ title: "hello" });
  await actions.createWithText({ title: "goodbye" });
  await actions.createWithText({ title: "hello" });
  await actions.createWithText({ title: "goodbye", isActive: false });
  await actions.createWithText({ title: null, isActive: false });

  expect(
    await actions.listWithTextPermissionFromExplicitInput({
      where: {
        isActive: { equals: true },
        explTitle: "hello",
      },
    })
  ).toHaveAuthorizationError();
});

test("permission on explicit input - one not matching null value - is not authorized", async () => {
  await actions.createWithText({ title: "hello" });
  await actions.createWithText({ title: null });
  await actions.createWithText({ title: "hello" });
  await actions.createWithText({ title: "goodbye", isActive: false });
  await actions.createWithText({ title: null, isActive: false });

  expect(
    await actions.listWithTextPermissionFromExplicitInput({
      where: {
        isActive: { equals: true },
        explTitle: "hello",
      },
    })
  ).toHaveAuthorizationError();
});
