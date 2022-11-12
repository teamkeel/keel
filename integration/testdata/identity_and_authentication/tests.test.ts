import { test, expect, actions, Post, Identity } from "@teamkeel/testing";

test("create identity", async () => {
  const { identityId, identityCreated } = await actions.authenticate({
    createIfNotExists: true,
    email: "user@keel.xyz",
    password: "1234",
  });

  expect(identityCreated).toEqual(true);
});

test("authenticate - invalid email - respond with invalid email address error", async () => {
  expect(
    await actions.authenticate({
      createIfNotExists: true,
      email: "user",
      password: "1234",
    })
  ).toHaveError({
    message: "invalid email address",
  });
});

test("authenticate - empty password - respond with password cannot be empty error", async () => {
  expect(
    await actions.authenticate({
      createIfNotExists: true,
      email: "user@keel.xyz",
      password: "",
    })
  ).toHaveError({
    message: "password cannot be empty",
  });
});

test("authenticate - new identity and createIfNotExists false - do not create identity", async () => {
  const { identityId, identityCreated } = await actions.authenticate({
    createIfNotExists: false,
    email: "user@keel.xyz",
    password: "1234",
  });

  expect(identityId).toBeEmpty();
  expect(identityCreated).toEqual(false);
});

test("authenticate - new identity - new identity created", async () => {
  const { identityId: id, identityCreated: created } =
    await actions.authenticate({
      createIfNotExists: true,
      email: "user@keel.xyz",
      password: "1234",
    });

  expect(id).notToBeEmpty();
  expect(created).toEqual(true);
});

test("authenticate - existing identity - authenticated", async () => {
  const { identityId: id1, identityCreated: created1 } =
    await actions.authenticate({
      createIfNotExists: true,
      email: "user@keel.xyz",
      password: "1234",
    });

  const { identityId: id2, identityCreated: created2 } =
    await actions.authenticate({
      createIfNotExists: true,
      email: "user@keel.xyz",
      password: "1234",
    });

  expect(id1).toEqual(id2);
  expect(created1).toEqual(true);
  expect(created2).toEqual(false);
});

test("authenticate - incorrect credentials with existing identity - not authenticated", async () => {
  const {
    identityId: id1,
    identityCreated: created1,
    errors: errors1,
  } = await actions.authenticate({
    createIfNotExists: true,
    email: "user@keel.xyz",
    password: "1234",
  });

  const { identityId: id2, identityCreated: created2 } =
    await actions.authenticate({
      createIfNotExists: true,
      email: "user@keel.xyz",
      password: "zzzz",
    });

  var notEqualIdentities = id1 != id2;
  expect(notEqualIdentities).toEqual(true);
  expect(created1).toEqual(true);
  expect(created2).toEqual(false);
});

test("identity context permission - correct identity - permission satisfied", async () => {
  const { identityId } = await actions.authenticate({
    createIfNotExists: true,
    email: "user@keel.xyz",
    password: "1234",
  });

  const { object: identity } = await Identity.findOne({ id: identityId });

  const { object: post } = await actions
    .withIdentity(identity)
    .createPostWithIdentity({ title: "temp" });

  expect(
    await actions
      .withIdentity(identity)
      .getPostRequiresIdentity({ id: post.id })
  ).notToHaveAuthorizationError();
});

test("identity context permission - incorrect identity - permission not satisfied", async () => {
  const { identityId: id1 } = await actions.authenticate({
    createIfNotExists: true,
    email: "user1@keel.xyz",
    password: "1234",
  });

  const { identityId: id2 } = await actions.authenticate({
    createIfNotExists: true,
    email: "user2@keel.xyz",
    password: "1234",
  });

  const { object: identity1 } = await Identity.findOne({ id: id1 });
  const { object: identity2 } = await Identity.findOne({ id: id2 });

  const { object: post } = await actions
    .withIdentity(identity1)
    .createPostWithIdentity({ title: "temp" });

  expect(
    await actions
      .withIdentity(identity2)
      .getPostRequiresIdentity({ id: post.id })
  ).toHaveAuthorizationError();
});

test("isAuthenticated context permission - authenticated - permission satisfied", async () => {
  const { identityId } = await actions.authenticate({
    createIfNotExists: true,
    email: "user@keel.xyz",
    password: "1234",
  });

  const { object: identity } = await Identity.findOne({ id: identityId });

  const { object: post } = await actions
    .withIdentity(identity)
    .createPostWithIdentity({ title: "temp" });

  expect(
    await actions
      .withIdentity(identity)
      .getPostRequiresAuthentication({ id: post.id })
  ).notToHaveAuthorizationError();
});

test("isAuthenticated context permission - not authenticated - permission not satisfied", async () => {
  const { identityId } = await actions.authenticate({
    createIfNotExists: true,
    email: "user@keel.xyz",
    password: "1234",
  });

  const { object: identity } = await Identity.findOne({ id: identityId });

  const { object: post } = await actions
    .withIdentity(identity)
    .createPostWithIdentity({ title: "temp" });

  expect(
    await actions.getPostRequiresAuthentication({ id: post.id })
  ).toHaveAuthorizationError();
});

test("not isAuthenticated context permission - authenticated - permission satisfied", async () => {
  const { identityId } = await actions.authenticate({
    createIfNotExists: true,
    email: "user@keel.xyz",
    password: "1234",
  });

  const { object: identity } = await Identity.findOne({ id: identityId });

  const { object: post } = await actions
    .withIdentity(identity)
    .createPostWithIdentity({ title: "temp" });

  expect(
    await actions
      .withIdentity(identity)
      .getPostRequiresNoAuthentication({ id: post.id })
  ).toHaveAuthorizationError();
});

test("not isAuthenticated context permission - not authenticated - permission not satisfied", async () => {
  const { identityId } = await actions.authenticate({
    createIfNotExists: true,
    email: "user@keel.xyz",
    password: "1234",
  });

  const { object: identity } = await Identity.findOne({ id: identityId });

  const { object: post } = await actions
    .withIdentity(identity)
    .createPostWithIdentity({ title: "temp" });

  expect(
    await actions.getPostRequiresNoAuthentication({ id: post.id })
  ).notToHaveAuthorizationError();
});

test("isAuthenticated context set - authenticated - is set to true", async () => {
  const { identityId } = await actions.authenticate({
    createIfNotExists: true,
    email: "user@keel.xyz",
    password: "1234",
  });

  const { object: identity } = await Identity.findOne({ id: identityId });

  const { object: post } = await actions
    .withIdentity(identity)
    .createPostSetIsAuthenticated({ title: "temp" });

  expect(post.isAuthenticated).toEqual(true);
});

test("isAuthenticated context set - not authenticated - is set to false", async () => {
  const { object: post } = await actions.createPostSetIsAuthenticated({
    title: "temp",
  });

  expect(post.isAuthenticated).toEqual(false);
});

// todo:  permission test against null.  Requires this fix:  https://linear.app/keel/issue/DEV-195/permissions-support-null-operand-with-identity-type

// todo:  permission test against another identity field.  Requires this fix: https://linear.app/keel/issue/DEV-196/permissions-support-identity-type-operand-with-identity-comparison
