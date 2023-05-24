import { test, expect, beforeEach } from "vitest";
import { actions, resetDatabase } from "@teamkeel/testing";
import { PostType } from "@teamkeel/sdk";
import { isNullishCoalesce } from "typescript";

beforeEach(resetDatabase);

test("string permission on literal - all matching - is authorized", async () => {
  await actions.createWithText({ title: { value: "hello" } });
  await actions.createWithText({ title: { value: "hello" } });
  await actions.createWithText({ title: { value: "hello" } });
  await actions.createWithText({
    title: { value: "goodbye" },
    isActive: false,
  });
  await actions.createWithText({ title: { isNull: true }, isActive: false });

  const r = await actions.listWithTextPermissionLiteral({
    where: {
      isActive: { equals: true },
    },
  });
  expect(r.results.length).toEqual(3);
});

test("string permission on literal - one not matching value - is not authorized", async () => {
  await actions.createWithText({ title: { value: "hello" } });
  await actions.createWithText({ title: { value: "goodbye" } });
  await actions.createWithText({ title: { value: "hello" } });

  await expect(
    actions.listWithTextPermissionLiteral({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("string permission on literal - one not matching null value - is not authorized", async () => {
  await actions.createWithText({ title: { value: "hello" } });
  await actions.createWithText({ title: { isNull: true } });
  await actions.createWithText({ title: { value: "hello" } });

  await expect(
    actions.listWithTextPermissionLiteral({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("number permission on literal - all matching - is authorized", async () => {
  await actions.createWithNumber({ views: { value: 1 } });
  await actions.createWithNumber({ views: { value: 1 } });
  await actions.createWithNumber({ views: { value: 1 } });
  await actions.createWithNumber({ views: { value: 100 }, isActive: false });
  await actions.createWithNumber({ views: { isNull: true }, isActive: false });

  const r = await actions.listWithNumberPermissionLiteral({
    where: {
      isActive: { equals: true },
    },
  });
  expect(r.results.length).toEqual(3);
});

test("number permission on literal - one not matching value - is not authorized", async () => {
  await actions.createWithNumber({ views: { value: 1 } });
  await actions.createWithNumber({ views: { value: 100 } });
  await actions.createWithNumber({ views: { value: 1 } });

  await expect(
    actions.listWithNumberPermissionLiteral({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("number permission on literal - one not matching null value - is not authorized", async () => {
  await actions.createWithNumber({ views: { value: 1 } });
  await actions.createWithNumber({ views: { isNull: true } });
  await actions.createWithNumber({ views: { value: 1 } });

  await expect(
    actions.listWithNumberPermissionLiteral({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("boolean permission on literal - all matching - is authorized", async () => {
  await actions.createWithBoolean({ active: { value: true } });
  await actions.createWithBoolean({ active: { value: true } });
  await actions.createWithBoolean({ active: { value: true } });
  await actions.createWithBoolean({
    active: { value: false },
    isActive: false,
  });
  await actions.createWithBoolean({
    active: { isNull: true },
    isActive: false,
  });

  const r = await actions.listWithBooleanPermissionLiteral({
    where: {
      isActive: { equals: true },
    },
  });
  expect(r.results.length).toEqual(3);
});

test("boolean permission on literal - one not matching value - is not authorized", async () => {
  await actions.createWithBoolean({ active: { value: true } });
  await actions.createWithBoolean({ active: { value: false } });
  await actions.createWithBoolean({ active: { value: true } });

  await expect(
    actions.listWithBooleanPermissionLiteral({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("boolean permission on literal - one not matching null value - is not authorized", async () => {
  await actions.createWithBoolean({ active: { value: true } });
  await actions.createWithBoolean({ active: { isNull: true } });
  await actions.createWithBoolean({ active: { value: true } });

  await expect(
    actions.listWithBooleanPermissionLiteral({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("enum permission on literal - all matching - is authorized", async () => {
  await actions.createWithEnum({ type: { value: PostType.Technical } });
  await actions.createWithEnum({ type: { value: PostType.Technical } });
  await actions.createWithEnum({ type: { value: PostType.Technical } });
  await actions.createWithEnum({
    type: { value: PostType.Food },
    isActive: false,
  });
  await actions.createWithEnum({ type: { isNull: true }, isActive: false });

  const r = await actions.listWithEnumPermissionLiteral({
    where: {
      isActive: { equals: true },
    },
  });
  expect(r.results.length).toEqual(3);
});

test("enum permission on literal - one not matching value - is not authorized", async () => {
  await actions.createWithEnum({ type: { value: PostType.Technical } });
  await actions.createWithEnum({ type: { value: PostType.Food } });
  await actions.createWithEnum({ type: { value: PostType.Technical } });

  await expect(
    actions.listWithEnumPermissionLiteral({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("enum permission on literal - one not matching null value - is not authorized", async () => {
  await actions.createWithEnum({ type: { value: PostType.Technical } });
  await actions.createWithEnum({ type: { isNull: true } });
  await actions.createWithEnum({ type: { value: PostType.Technical } });

  await expect(
    actions.listWithEnumPermissionLiteral({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("string permission on field - all matching - is authorized", async () => {
  await actions.createWithText({ title: { value: "hello" } });
  await actions.createWithText({ title: { value: "hello" } });
  await actions.createWithText({ title: { value: "hello" } });
  await actions.createWithText({
    title: { value: "goodbye" },
    isActive: false,
  });
  await actions.createWithText({ title: { isNull: true }, isActive: false });

  const r = await actions.listWithTextPermissionFromField({
    where: {
      isActive: { equals: true },
    },
  });
  expect(r.results.length).toEqual(3);
});

test("string permission on field - one not matching value - is not authorized", async () => {
  await actions.createWithText({ title: { value: "hello" } });
  await actions.createWithText({ title: { value: "goodbye" } });
  await actions.createWithText({ title: { value: "hello" } });

  await expect(
    actions.listWithTextPermissionFromField({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("string permission on field - one not matching null value - is not authorized", async () => {
  await actions.createWithText({ title: { value: "hello" } });
  await actions.createWithText({ title: { isNull: true } });
  await actions.createWithText({ title: { value: "hello" } });

  await expect(
    actions.listWithTextPermissionFromField({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("number permission on field - all matching - is authorized", async () => {
  await actions.createWithNumber({ views: { value: 1 } });
  await actions.createWithNumber({ views: { value: 1 } });
  await actions.createWithNumber({ views: { value: 1 } });
  await actions.createWithNumber({ views: { value: 100 }, isActive: false });
  await actions.createWithNumber({ views: { isNull: true }, isActive: false });

  const r = await actions.listWithNumberPermissionFromField({
    where: {
      isActive: { equals: true },
    },
  });
  expect(r.results.length).toEqual(3);
});

test("number permission on field - one not matching value - is not authorized", async () => {
  await actions.createWithNumber({ views: { value: 1 } });
  await actions.createWithNumber({ views: { value: 100 } });
  await actions.createWithNumber({ views: { value: 1 } });

  await expect(
    actions.listWithNumberPermissionFromField({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("number permission on field - one not matching null value - is not authorized", async () => {
  await actions.createWithNumber({ views: { value: 1 } });
  await actions.createWithNumber({ views: { isNull: true } });
  await actions.createWithNumber({ views: { value: 1 } });

  await expect(
    actions.listWithNumberPermissionFromField({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("boolean permission on field - all matching - is authorized", async () => {
  await actions.createWithBoolean({ active: { value: true } });
  await actions.createWithBoolean({ active: { value: true } });
  await actions.createWithBoolean({ active: { value: true } });
  await actions.createWithBoolean({
    active: { value: false },
    isActive: false,
  });
  await actions.createWithBoolean({
    active: { isNull: true },
    isActive: false,
  });

  const r = await actions.listWithBooleanPermissionFromField({
    where: {
      isActive: { equals: true },
    },
  });
  expect(r.results.length).toEqual(3);
});

test("boolean permission on field - one not matching value - field is not authorized", async () => {
  await actions.createWithBoolean({ active: { value: true } });
  await actions.createWithBoolean({ active: { value: false } });
  await actions.createWithBoolean({ active: { value: true } });

  await expect(
    actions.listWithBooleanPermissionFromField({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("boolean permission on field - one not matching null value - is not authorized", async () => {
  await actions.createWithBoolean({ active: { value: true } });
  await actions.createWithBoolean({ active: { isNull: true } });
  await actions.createWithBoolean({ active: { value: true } });

  await expect(
    actions.listWithBooleanPermissionFromField({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("enum permission on field name - all matching - is authorized", async () => {
  await actions.createWithEnum({ type: { value: PostType.Technical } });
  await actions.createWithEnum({ type: { value: PostType.Technical } });
  await actions.createWithEnum({ type: { value: PostType.Technical } });
  await actions.createWithEnum({
    type: { value: PostType.Food },
    isActive: false,
  });
  await actions.createWithEnum({ type: { isNull: true }, isActive: false });

  await expect(
    actions.listWithEnumPermissionFromField({
      where: {
        isActive: { equals: true },
      },
    })
  ).not.toHaveAuthorizationError();
});

test("enum permission on field name - one not matching value - is not authorized", async () => {
  await actions.createWithEnum({ type: { value: PostType.Technical } });
  await actions.createWithEnum({ type: { value: PostType.Food } });
  await actions.createWithEnum({ type: { value: PostType.Technical } });

  await expect(
    actions.listWithEnumPermissionFromField({
      where: {
        isActive: { equals: true },
      },
    })
  ).toHaveAuthorizationError();
});

test("enum permission on field name - one not matching null value - is not authorized", async () => {
  await actions.createWithEnum({ type: { value: PostType.Technical } });
  await actions.createWithEnum({ type: { isNull: true } });
  await actions.createWithEnum({ type: { value: PostType.Technical } });

  await expect(
    actions.listWithEnumPermissionFromField({
      where: {
        isActive: { equals: true },
      },
    })
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

  await actions.withAuthToken(token).createWithIdentity({});

  await actions.withAuthToken(token).createWithIdentity({});

  const { token: token2 } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "anotheruser@keel.xyz",
      password: "beep",
    },
  });

  await actions.withAuthToken(token2).createWithIdentity({ isActive: false });

  await expect(
    actions
      .withAuthToken(token)
      .listWithIdentityPermission({ where: { isActive: { equals: true } } })
  ).not.toHaveAuthorizationError();
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

  await actions.withAuthToken(token).createWithIdentity({});
  await actions.withAuthToken(token2).createWithIdentity({});

  await expect(
    actions.withAuthToken(token2).listWithIdentityPermission({
      where: {
        isActive: { equals: true },
      },
    })
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

  await actions.withAuthToken(token).createWithIdentity({});
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
  await actions.createWithText({ title: { value: "hello" } });

  await expect(
    actions.listWithTrueValuePermission({})
  ).not.toHaveAuthorizationError();
});
