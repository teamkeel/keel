import { actions, models, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";
import { PostType } from "@teamkeel/sdk";

beforeEach(resetDatabase);

test("string permission on literal - un matching value - is not authorized", async () => {
  const post = await actions.createWithText({ title: "hello" });

  await expect(
    actions.updateWithTextPermissionLiteral({
      where: { id: post.id },
      values: { title: "goodbye" },
    })
  ).toHaveAuthorizationError();

  // Ensure the update transaction did rollback
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.title).equals("hello");
});

test("string permission on literal - null value - is not authorized", async () => {
  const post = await actions.createWithText({ title: "goodbye" });

  await expect(
    actions.updateWithTextPermissionLiteral({
      where: { id: post.id },
      values: { title: null },
    })
  ).toHaveAuthorizationError();

  // Ensure the update transaction did rollback
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.title).equals("goodbye");
});

test("number permission on literal - unmatching value - is not authorized", async () => {
  const post = await actions.createWithNumber({ views: 1 });

  await expect(
    actions.updateWithNumberPermissionLiteral({
      where: { id: post.id },
      values: { views: 100 },
    })
  ).toHaveAuthorizationError();

  // Ensure the update transaction did rollback
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.views).equals(1);
});

test("number permission on literal - null value - is not authorized", async () => {
  const post = await actions.createWithNumber({ views: 100 });

  await expect(
    actions.updateWithNumberPermissionLiteral({
      where: { id: post.id },
      values: { views: null },
    })
  ).toHaveAuthorizationError();

  // Ensure the update transaction did rollback
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.views).equals(100);
});

test("boolean permission on literal - unmatching value - is not authorized", async () => {
  const post = await actions.createWithBoolean({ active: true });

  await expect(
    actions.updateWithBooleanPermissionLiteral({
      where: { id: post.id },
      values: { active: false },
    })
  ).toHaveAuthorizationError();

  // Ensure the update transaction did rollback
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.active).equals(true);
});

test("boolean permission on literal - null value - is not authorized", async () => {
  const post = await actions.createWithBoolean({ active: false });

  await expect(
    actions.updateWithBooleanPermissionLiteral({
      where: { id: post.id },
      values: { active: null },
    })
  ).toHaveAuthorizationError();

  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.active).equals(false);
});

test("enum permission on literal - unmatching value - is not authorized", async () => {
  const post = await actions.createWithEnum({ type: PostType.Technical });

  await expect(
    actions.updateWithEnumPermissionLiteral({
      where: { id: post.id },
      values: { type: PostType.Lifestyle },
    })
  ).toHaveAuthorizationError();

  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.type).equals(PostType.Technical);
});

test("enum permission on literal - null value - is not authorized", async () => {
  const post = await actions.createWithEnum({ type: null });

  await expect(
    actions.updateWithEnumPermissionLiteral({
      where: { id: post.id },
      values: { type: null },
    })
  ).toHaveAuthorizationError();

  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.type).equals(null);
});

test("string permission on field - unmatching value - is not authorized", async () => {
  const post = await actions.createWithText({ title: "hello" });

  await expect(
    actions.updateWithTextPermissionFromField({
      where: { id: post.id },
      values: { title: "goodbye" },
    })
  ).toHaveAuthorizationError();

    // Ensure the update transaction did rollback
    const samePost = await models.post.findOne({ id: post.id });
    expect(samePost!.title).equals("hello");
});

test("string permission on field - null value - is not authorized", async () => {
  const post = await actions.createWithText({ title: "goodbye" });

  await expect(
    actions.updateWithTextPermissionFromField({
      where: { id: post.id },
      values: { title: null },
    })
  ).toHaveAuthorizationError();

    // Ensure the update transaction did rollback
    const samePost = await models.post.findOne({ id: post.id });
    expect(samePost!.title).equals("goodbye");
});

test("number permission on field - unmatching value - is not authorized", async () => {
  const post = await actions.createWithNumber({ views: 1 });

  await expect(
    actions.updateWithNumberPermissionFromField({
      where: { id: post.id },
      values: { views: 100 },
    })
  ).toHaveAuthorizationError();

    // Ensure the update transaction did rollback
    const samePost = await models.post.findOne({ id: post.id });
    expect(samePost!.views).equals(1);
});

test("number permission on field - null value - is not authorized", async () => {
  const post = await actions.createWithNumber({ views: 100 });

  await expect(
    actions.updateWithNumberPermissionFromField({
      where: { id: post.id },
      values: { views: null },
    })
  ).toHaveAuthorizationError();

      // Ensure the update transaction did rollback
      const samePost = await models.post.findOne({ id: post.id });
      expect(samePost!.views).equals(100);
});

test("boolean permission on field - unmatching value - is not authorized", async () => {
  const post = await actions.createWithBoolean({ active: true });

  await expect(
    actions.updateWithBooleanPermissionFromField({
      where: { id: post.id },
      values: { active: false },
    })
  ).toHaveAuthorizationError();

      // Ensure the update transaction did rollback
      const samePost = await models.post.findOne({ id: post.id });
      expect(samePost!.active).equals(true);
});

test("boolean permission on field - null - is not authorized", async () => {
  const post = await actions.createWithBoolean({ active: false });

  await expect(
    actions.updateWithBooleanPermissionFromField({
      where: { id: post.id },
      values: { active: null },
    })
  ).toHaveAuthorizationError();

     // Ensure the update transaction did rollback
     const samePost = await models.post.findOne({ id: post.id });
     expect(samePost!.active).equals(false);
});

test("enum permission on field - unmatching value - is not authorized", async () => {
  const post = await actions.createWithEnum({ type: PostType.Technical });

  await expect(
    actions.updateWithEnumPermissionFromField({
      where: { id: post.id },
      values: { type: PostType.Lifestyle },
    })
  ).toHaveAuthorizationError();

     // Ensure the update transaction did rollback
     const samePost = await models.post.findOne({ id: post.id });
     expect(samePost!.type).equals(PostType.Technical);
});

test("enum permission on field - null - is not authorized", async () => {
  const post = await actions.createWithEnum({ type: null });

  await expect(
    actions.updateWithEnumPermissionFromField({
      where: { id: post.id },
      values: { type: null },
    })
  ).toHaveAuthorizationError();

   // Ensure the update transaction did rollback
   const samePost = await models.post.findOne({ id: post.id });
   expect(samePost!.type).equals(null);
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

  await expect(
    actions.withAuthToken(token).updateWithIdentityPermission({
      where: { id: post.id },
      values: { title: "hello" },
    })
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

  const post = await actions.withAuthToken(token).createWithIdentity({});

  await expect(
    actions.withAuthToken(token2).updateWithIdentityPermission({
      where: { id: post.id },
      values: { title: "hello" },
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

  const post = await actions.withAuthToken(token).createWithIdentity({});

  await expect(
    actions.updateWithIdentityPermission({
      where: { id: post.id },
      values: { title: "hello" },
    })
  ).toHaveAuthorizationError();
});

test("true value permission - unauthenticated identity - is authorized", async () => {
  const post = await actions.createWithText({ title: "hello" });

  await expect(
    actions.updateWithTrueValuePermission({
      where: { id: post.id },
      values: { title: "hello again" },
    })
  ).not.toHaveAuthorizationError();
});
