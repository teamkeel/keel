import { actions, models, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";
import { PostType } from "@teamkeel/sdk";

beforeEach(resetDatabase);

test("string permission on literal - matching value - is authorized", async () => {
  const post = await actions.createWithText({ title: "hello" });

  await expect(
    actions.updateWithTextPermissionLiteral({
      where: { id: post.id },
      values: { title: "goodbye" },
    })
  ).not.toHaveAuthorizationError();

  // Ensure the update completed
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.title).equals("goodbye");
});

test("string permission on literal - not matching value - is not authorized", async () => {
  const post = await actions.createWithText({ title: "goodbye" });

  await expect(
    actions.updateWithTextPermissionLiteral({
      where: { id: post.id },
      values: { title: "hello" },
    })
  ).toHaveAuthorizationError();

  // Ensure the update did not complete
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.title).equals("goodbye");
});

test("string permission on literal - null value - is not authorized", async () => {
  const post = await actions.createWithText({ title: "goodbye" });

  await expect(
    actions.updateWithTextPermissionLiteral({
      where: { id: post.id },
      values: { title: null },
    })
  ).toHaveAuthorizationError();

  // Ensure the update did not complete
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.title).equals("goodbye");
});

test("number permission on literal - matching value - is authorized", async () => {
  const post = await actions.createWithNumber({ views: 1 });

  await expect(
    actions.updateWithNumberPermissionLiteral({
      where: { id: post.id },
      values: { views: 100 },
    })
  ).not.toHaveAuthorizationError();

  // Ensure the update completed
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.views).equals(100);
});

test("number permission on literal - not matching value - is not authorized", async () => {
  const post = await actions.createWithNumber({ views: 100 });

  await expect(
    actions.updateWithNumberPermissionLiteral({
      where: { id: post.id },
      values: { views: 1 },
    })
  ).toHaveAuthorizationError();

  // Ensure the update did not complete
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.views).equals(100);
});

test("number permission on literal - null value - is not authorized", async () => {
  const post = await actions.createWithNumber({ views: 100 });

  await expect(
    actions.updateWithNumberPermissionLiteral({
      where: { id: post.id },
      values: { views: null },
    })
  ).toHaveAuthorizationError();

  // Ensure the update did not complete
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.views).equals(100);
});

test("boolean permission on literal - matching value - is authorized", async () => {
  const post = await actions.createWithBoolean({ active: true });

  await expect(
    actions.updateWithBooleanPermissionLiteral({
      where: { id: post.id },
      values: { active: false },
    })
  ).not.toHaveAuthorizationError();

  // Ensure the update completed
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.active).equals(false);
});

test("boolean permission on literal - not matching value - is not authorized", async () => {
  const post = await actions.createWithBoolean({ active: false });

  await expect(
    actions.updateWithBooleanPermissionLiteral({
      where: { id: post.id },
      values: { active: true },
    })
  ).toHaveAuthorizationError();

  // Ensure the update did not complete
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.active).equals(false);
});

test("boolean permission on literal - null value - is not authorized", async () => {
  const post = await actions.createWithBoolean({ active: false });

  await expect(
    actions.updateWithBooleanPermissionLiteral({
      where: { id: post.id },
      values: { active: null },
    })
  ).toHaveAuthorizationError();

  // Ensure the update did not complete
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.active).equals(false);
});

test("enum permission on literal - matching value - is authorized", async () => {
  const post = await actions.createWithEnum({ type: PostType.Technical });

  await expect(
    actions.updateWithEnumPermissionLiteral({
      where: { id: post.id },
      values: { type: PostType.Lifestyle },
    })
  ).not.toHaveAuthorizationError();

  // Ensure the update completed
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.type).equals(PostType.Lifestyle);
});

test("enum permission on literal - not matching value - is not authorized", async () => {
  const post = await actions.createWithEnum({ type: PostType.Lifestyle });

  await expect(
    actions.updateWithEnumPermissionLiteral({
      where: { id: post.id },
      values: { type: PostType.Lifestyle },
    })
  ).toHaveAuthorizationError();

  // Ensure the update did not complete
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.type).equals(PostType.Lifestyle);
});

test("enum permission on literal - null value - is not authorized", async () => {
  const post = await actions.createWithEnum({ type: null });

  await expect(
    actions.updateWithEnumPermissionLiteral({
      where: { id: post.id },
      values: { type: null },
    })
  ).toHaveAuthorizationError();

  // Ensure the update did not complete
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.type).equals(null);
});

test("string permission on field - matching value - is authorized", async () => {
  const post = await actions.createWithText({ title: "hello" });

  await expect(
    actions.updateWithTextPermissionFromField({
      where: { id: post.id },
      values: { title: "goodbye" },
    })
  ).not.toHaveAuthorizationError();

  // Ensure the update complete
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.title).equals("goodbye");
});

test("string permission on field - not matching value - is not authorized", async () => {
  const post = await actions.createWithText({ title: "goodbye" });

  await expect(
    actions.updateWithTextPermissionFromField({
      where: { id: post.id },
      values: { title: "hello" },
    })
  ).toHaveAuthorizationError();

  // Ensure the update did not complete
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.title).equals("goodbye");
});

test("string permission on field - null value - is not authorized", async () => {
  const post = await actions.createWithText({ title: "goodbye" });

  await expect(
    actions.updateWithTextPermissionFromField({
      where: { id: post.id },
      values: { title: null },
    })
  ).toHaveAuthorizationError();

  // Ensure the update did not complete
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.title).equals("goodbye");
});

test("number permission on field - matching value - is authorized", async () => {
  const post = await actions.createWithNumber({ views: 1 });

  await expect(
    actions.updateWithNumberPermissionFromField({
      where: { id: post.id },
      values: { views: 100 },
    })
  ).not.toHaveAuthorizationError();

  // Ensure the update completed
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.views).equals(100);
});

test("number permission on field - not matching value - is not authorized", async () => {
  const post = await actions.createWithNumber({ views: 100 });

  await expect(
    actions.updateWithNumberPermissionFromField({
      where: { id: post.id },
      values: { views: 1 },
    })
  ).toHaveAuthorizationError();

  // Ensure the update did not complete
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.views).equals(100);
});

test("number permission on field - null value - is not authorized", async () => {
  const post = await actions.createWithNumber({ views: 100 });

  await expect(
    actions.updateWithNumberPermissionFromField({
      where: { id: post.id },
      values: { views: null },
    })
  ).toHaveAuthorizationError();

  // Ensure the update did not complete
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.views).equals(100);
});

test("boolean permission on field - matching value - is authorized", async () => {
  const post = await actions.createWithBoolean({ active: true });

  await expect(
    actions.updateWithBooleanPermissionFromField({
      where: { id: post.id },
      values: { active: false },
    })
  ).not.toHaveAuthorizationError();

  // Ensure the update completed
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.active).equals(false);
});

test("boolean permission on field - field is not authorized", async () => {
  const post = await actions.createWithBoolean({ active: false });

  await expect(
    actions.updateWithBooleanPermissionFromField({
      where: { id: post.id },
      values: { active: true },
    })
  ).toHaveAuthorizationError();

  // Ensure the update did not complete
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.active).equals(false);
});

test("boolean permission on field - null - is not authorized", async () => {
  const post = await actions.createWithBoolean({ active: false });

  await expect(
    actions.updateWithBooleanPermissionFromField({
      where: { id: post.id },
      values: { active: null },
    })
  ).toHaveAuthorizationError();

  // Ensure the update did not complete
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.active).equals(false);
});

test("enum permission on field - matching value - is authorized", async () => {
  const post = await actions.createWithEnum({ type: PostType.Technical });

  await expect(
    actions.updateWithEnumPermissionFromField({
      where: { id: post.id },
      values: { type: PostType.Lifestyle },
    })
  ).not.toHaveAuthorizationError();

  // Ensure the update completed
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.type).equals(PostType.Lifestyle);
});

test("enum permission on field - field is not authorized", async () => {
  const post = await actions.createWithEnum({ type: PostType.Lifestyle });

  await expect(
    actions.updateWithEnumPermissionFromField({
      where: { id: post.id },
      values: { type: PostType.Technical },
    })
  ).toHaveAuthorizationError();

  // Ensure the update did not complete
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.type).equals(PostType.Lifestyle);
});

test("enum permission on field - null - is not authorized", async () => {
  const post = await actions.createWithEnum({ type: null });

  await expect(
    actions.updateWithEnumPermissionFromField({
      where: { id: post.id },
      values: { type: null },
    })
  ).toHaveAuthorizationError();

  // Ensure the update did not complete
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.type).equals(null);
});

test("identity permission - correct identity in context - is authorized", async () => {
  const identity = await models.identity.create({
    email: "user@keel.xyz",
    issuer: "https://keel.so",
  });

  const post = await actions.withIdentity(identity).createWithIdentity({});

  await expect(
    actions.withIdentity(identity).updateWithIdentityPermission({
      where: { id: post.id },
      values: { title: "hello" },
    })
  ).not.toHaveAuthorizationError();
});

test("identity permission - incorrect identity in context - is not authorized", async () => {
  const identity1 = await models.identity.create({
    email: "user1@keel.xyz",
    issuer: "https://keel.so",
  });

  const identity2 = await models.identity.create({
    email: "user2@keel.xyz",
    issuer: "https://keel.so",
  });

  const post = await actions.withIdentity(identity1).createWithIdentity({});

  await expect(
    actions.withIdentity(identity2).updateWithIdentityPermission({
      where: { id: post.id },
      values: { title: "hello" },
    })
  ).toHaveAuthorizationError();
});

test("identity permission - no identity in context - is not authorized", async () => {
  const identity = await models.identity.create({
    email: "user@keel.xyz",
    issuer: "https://keel.so",
  });

  const post = await actions.withIdentity(identity).createWithIdentity({});

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
