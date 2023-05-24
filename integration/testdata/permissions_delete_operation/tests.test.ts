import { resetDatabase, actions, models } from "@teamkeel/testing";
import { PostType } from "@teamkeel/sdk";
import { beforeEach, test, expect } from "vitest";

beforeEach(resetDatabase);

test("string permission on literal - matching value - is authorized", async () => {
  const post = await actions.createWithTitle({ title: { value: "hello" } });

  const deleted = await actions.deleteWithTextPermissionLiteral({
    id: post.id,
  });
  expect(deleted).toEqual(post.id);
});

test("string permission on literal - not matching value - is not authorized", async () => {
  const post = await actions.createWithTitle({
    title: { value: "not hello" },
  });

  await expect(
    actions.deleteWithTextPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();

  expect(await models.post.findOne({ id: post.id })).not.toBeNull();
});

test("string permission on literal - not matching null value - is not authorized", async () => {
  const post = await actions.createWithTitle({ title: { isNull: true } });

  await expect(
    actions.deleteWithTextPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();

  expect(await models.post.findOne({ id: post.id })).not.toBeNull();
});

test("number permission on literal - matching value - is authorized", async () => {
  const post = await actions.createWithViews({ views: { value: 5 } });

  const deleted = await actions.deleteWithNumberPermissionLiteral({
    id: post.id,
  });
  expect(deleted).toEqual(post.id);

  expect(await models.post.findOne({ id: post.id })).toBeNull();
});

test("number permission on literal - not matching value - is not authorized", async () => {
  const post = await actions.createWithViews({ views: { value: 500 } });

  await expect(
    actions.deleteWithNumberPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();

  expect(await models.post.findOne({ id: post.id })).not.toBeNull();
});

test("number permission on literal - not matching null value - is not authorized", async () => {
  const post = await actions.createWithViews({ views: { isNull: true } });

  await expect(
    actions.deleteWithNumberPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();

  expect(await models.post.findOne({ id: post.id })).not.toBeNull();
});

test("boolean permission on literal - matching value - is authorized", async () => {
  const post = await actions.createWithActive({ active: { value: true } });

  const deleted = await actions.deleteWithBooleanPermissionLiteral({
    id: post.id,
  });
  expect(deleted).toEqual(post.id);

  expect(await models.post.findOne({ id: post.id })).toBeNull();
});

test("boolean permission on literal - not matching value - is not authorized", async () => {
  const post = await actions.createWithActive({ active: { value: false } });

  await expect(
    actions.deleteWithBooleanPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();

  expect(await models.post.findOne({ id: post.id })).not.toBeNull();
});

test("boolean permission on literal - not matching null value - is not authorized", async () => {
  const post = await actions.createWithActive({ active: { isNull: true } });

  await expect(
    actions.deleteWithBooleanPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();

  expect(await models.post.findOne({ id: post.id })).not.toBeNull();
});

test("enum permission on literal - matching value - is authorized", async () => {
  const post = await actions.createWithEnum({
    type: { value: PostType.Technical },
  });

  const deleted = await actions.deleteWithEnumPermissionLiteral({
    id: post.id,
  });
  expect(deleted).toEqual(post.id);

  expect(await models.post.findOne({ id: post.id })).toBeNull();
});

test("enum permission on literal - not matching value - is not authorized", async () => {
  const post = await actions.createWithEnum({
    type: { value: PostType.Lifestyle },
  });

  await expect(
    actions.deleteWithEnumPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();

  expect(await models.post.findOne({ id: post.id })).not.toBeNull();
});

test("enum permission on literal - not matching null value - is not authorized", async () => {
  const post = await actions.createWithEnum({ type: { isNull: true } });

  await expect(
    actions.deleteWithEnumPermissionLiteral({ id: post.id })
  ).toHaveAuthorizationError();

  expect(await models.post.findOne({ id: post.id })).not.toBeNull();
});

test("string permission on field - matching value - is authorized", async () => {
  const post = await actions.createWithTitle({ title: { value: "hello" } });

  const deleted = await actions.deleteWithTextPermissionOnField({
    id: post.id,
  });
  expect(deleted).toEqual(post.id);

  expect(await models.post.findOne({ id: post.id })).toBeNull();
});

test("string permission on field - not matching value - is not authorized", async () => {
  const post = await actions.createWithTitle({
    title: { value: "not hello" },
  });

  await expect(
    actions.deleteWithTextPermissionOnField({ id: post.id })
  ).toHaveAuthorizationError();

  expect(await models.post.findOne({ id: post.id })).not.toBeNull();
});

test("string permission on field - not matching null value - is not authorized", async () => {
  const post = await actions.createWithTitle({ title: { isNull: true } });

  await expect(
    actions.deleteWithTextPermissionOnField({ id: post.id })
  ).toHaveAuthorizationError();

  expect(await models.post.findOne({ id: post.id })).not.toBeNull();
});

test("number permission on field - matching value - is authorized", async () => {
  const post = await actions.createWithViews({ views: { value: 5 } });

  const deleted = await actions.deleteWithNumberPermissionOnField({
    id: post.id,
  });
  expect(deleted).toEqual(post.id);

  expect(await models.post.findOne({ id: post.id })).toBeNull();
});

test("number permission on field - not matching value - is not authorized", async () => {
  const post = await actions.createWithViews({ views: { value: 500 } });

  await expect(
    actions.deleteWithNumberPermissionOnField({ id: post.id })
  ).toHaveAuthorizationError();

  expect(await models.post.findOne({ id: post.id })).not.toBeNull();
});

test("number permission on field - not matching null value - is not authorized", async () => {
  const post = await actions.createWithViews({ views: { isNull: true } });

  await expect(
    actions.deleteWithNumberPermissionOnField({ id: post.id })
  ).toHaveAuthorizationError();

  expect(await models.post.findOne({ id: post.id })).not.toBeNull();
});

test("boolean permission on field - matching value - is authorized", async () => {
  const post = await actions.createWithActive({ active: { value: true } });

  const deleted = await actions.deleteWithBooleanPermissionOnField({
    id: post.id,
  });
  expect(deleted).toEqual(post.id);

  expect(await models.post.findOne({ id: post.id })).toBeNull();
});

test("boolean permission on field - not matching value - is not authorized", async () => {
  const post = await actions.createWithActive({ active: { value: false } });

  await expect(
    actions.deleteWithBooleanPermissionOnField({ id: post.id })
  ).toHaveAuthorizationError();

  expect(await models.post.findOne({ id: post.id })).not.toBeNull();
});

test("boolean permission on field - not matching null value - is not authorized", async () => {
  const post = await actions.createWithActive({ active: { isNull: true } });

  await expect(
    actions.deleteWithBooleanPermissionOnField({ id: post.id })
  ).toHaveAuthorizationError();

  expect(await models.post.findOne({ id: post.id })).not.toBeNull();
});

test("enum permission on field - matching value - is authorized", async () => {
  const post = await actions.createWithEnum({
    type: { value: PostType.Technical },
  });

  const deleted = await actions.deleteWithEnumPermissionOnField({
    id: post.id,
  });
  expect(deleted).toEqual(post.id);

  expect(await models.post.findOne({ id: post.id })).toBeNull();
});

test("enum permission on field - not matching value - is not authorized", async () => {
  const post = await actions.createWithEnum({
    type: { value: PostType.Lifestyle },
  });

  await expect(
    actions.deleteWithEnumPermissionOnField({ id: post.id })
  ).toHaveAuthorizationError();

  expect(await models.post.findOne({ id: post.id })).not.toBeNull();
});

test("enum permission on field - not matching null value - is not authorized", async () => {
  const post = await actions.createWithEnum({ type: { isNull: true } });

  await expect(
    actions.deleteWithEnumPermissionOnField({ id: post.id })
  ).toHaveAuthorizationError();

  expect(await models.post.findOne({ id: post.id })).not.toBeNull();
});

test("identity permission - correct identity in context - is authorized", async () => {
  const { token } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "user@keel.xyz",
      password: "1234",
    },
  });

  const post = await actions.withAuthToken(token).createWithIdentity({});

  const deleted = await actions
    .withAuthToken(token)
    .deleteWithRequiresSameIdentity({ id: post.id });

  expect(deleted).toEqual(post.id);
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

  const post = await actions.withAuthToken(token).createWithIdentity({});

  await expect(
    actions
      .withAuthToken(token2)
      .deleteWithRequiresSameIdentity({ id: post.id })
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

  const post = await actions.withAuthToken(token).createWithIdentity({});

  await expect(
    actions.deleteWithRequiresSameIdentity({ id: post.id })
  ).toHaveAuthorizationError();
});

test("true value permission - with unauthenticated identity - is authorized", async () => {
  const post = await actions.createWithActive({ active: { value: true } });

  const deleted = await actions.deleteWithTrueValuePermission({ id: post.id });
  expect(deleted).toEqual(post.id);
});
