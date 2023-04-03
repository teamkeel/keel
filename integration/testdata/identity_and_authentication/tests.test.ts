import { test, expect, expectTypeOf, beforeEach } from "vitest";
import { actions, models, resetDatabase } from "@teamkeel/testing";

beforeEach(resetDatabase);

test("create identity", async () => {
  const { identityCreated } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "user@keel.xyz",
      password: "1234",
    },
  });

  expect(identityCreated).toEqual(true);
});

test("authenticate - invalid email - respond with invalid email address error", async () => {
  await expect(
    actions.authenticate({
      createIfNotExists: true,
      emailPassword: {
        email: "user",
        password: "1234",
      },
    })
  ).rejects.toEqual({
    code: "ERR_INVALID_INPUT",
    message: "invalid email address",
  });
});

test("authenticate - empty password - respond with password cannot be empty error", async () => {
  await expect(
    actions.authenticate({
      createIfNotExists: true,
      emailPassword: {
        email: "user@keel.xyz",
        password: "",
      },
    })
  ).rejects.toEqual({
    code: "ERR_INVALID_INPUT",
    message: "password cannot be empty",
  });
});

test("authenticate - new identity and createIfNotExists false - respond with failed to authenticate error", async () => {
  await expect(
    actions.authenticate({
      createIfNotExists: false,
      emailPassword: {
        email: "user@keel.xyz",
        password: "1234",
      },
    })
  ).rejects.toEqual({
    code: "ERR_INVALID_INPUT",
    message: "failed to authenticate",
  });
});

test("authenticate - existing identity and createIfNotExists false - authenticated", async () => {
  const { identityCreated: created1 } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "user@keel.xyz",
      password: "1234",
    },
  });

  const { identityCreated: created2 } = await actions.authenticate({
    createIfNotExists: false,
    emailPassword: {
      email: "user@keel.xyz",
      password: "1234",
    },
  });

  const count = (await models.identity.findMany({})).length;

  expect(count).toEqual(1);
  expect(created1).toEqual(true);
  expect(created2).toEqual(false);
});

test("authenticate - new identity - new identity created", async () => {
  const authResponse = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "user@keel.xyz",
      password: "1234",
    },
  });
  expect(authResponse.token).toBeTruthy();
  expect(authResponse.identityCreated).toEqual(true);
});

test("authenticate - existing identity - authenticated", async () => {
  const { identityCreated: created1 } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "user@keel.xyz",
      password: "1234",
    },
  });

  const { identityCreated: created2 } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "user@keel.xyz",
      password: "1234",
    },
  });

  expect(created1).toEqual(true);
  expect(created2).toEqual(false);
});

test("authenticate - incorrect credentials with existing identity - not authenticated", async () => {
  const { identityCreated: created1 } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "user@keel.xyz",
      password: "1234",
    },
  });

  expect(created1).toEqual(true);

  await expect(
    actions.authenticate({
      createIfNotExists: true,
      emailPassword: {
        email: "user@keel.xyz",
        password: "zzzz",
      },
    })
  ).rejects.toEqual({
    code: "ERR_INVALID_INPUT",
    message: "failed to authenticate",
  });
});

test("identity context permission - correct identity - permission satisfied", async () => {
  const authResponse = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "user@keel.xyz",
      password: "1234",
    },
  });

  const authedActions = actions.withAuthToken(authResponse.token);

  const post = await authedActions.createPostWithIdentity({ title: "temp" });

  await expect(
    authedActions.getPostRequiresIdentity({ id: post.id })
  ).resolves.toEqual(post);
});

test("identity context permission - incorrect identity - permission not satisfied", async () => {
  const { token: token1 } = await actions.authenticate({
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

  const post = await actions
    .withAuthToken(token1)
    .createPostWithIdentity({ title: "temp" });

  await expect(
    actions.withAuthToken(token2).getPostRequiresIdentity({ id: post.id })
  ).toHaveAuthorizationError();
});

test("isAuthenticated context permission - authenticated - permission satisfied", async () => {
  const { token } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "user@keel.xyz",
      password: "1234",
    },
  });

  const post = await actions
    .withAuthToken(token)
    .createPostWithIdentity({ title: "temp" });

  await expect(
    actions.withAuthToken(token).getPostRequiresAuthentication({ id: post.id })
  ).resolves.toEqual(post);
});

test("isAuthenticated context permission - not authenticated - permission not satisfied", async () => {
  const { token } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "user@keel.xyz",
      password: "1234",
    },
  });

  const post = await actions
    .withAuthToken(token)
    .createPostWithIdentity({ title: "temp" });

  await expect(
    actions.getPostRequiresAuthentication({ id: post.id })
  ).toHaveAuthorizationError();
});

test("not isAuthenticated context permission - authenticated - permission satisfied", async () => {
  const { token } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "user@keel.xyz",
      password: "1234",
    },
  });

  const post = await actions
    .withAuthToken(token)
    .createPostWithIdentity({ title: "temp" });

  await expect(
    actions
      .withAuthToken(token)
      .getPostRequiresNoAuthentication({ id: post.id })
  ).toHaveAuthorizationError();
});

test("not isAuthenticated context permission - not authenticated - permission satisfied", async () => {
  const { token } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "user@keel.xyz",
      password: "1234",
    },
  });

  const post = await actions
    .withAuthToken(token)
    .createPostWithIdentity({ title: "temp" });

  await expect(
    actions.getPostRequiresNoAuthentication({ id: post.id })
  ).resolves.toEqual(post);
});

test("isAuthenticated context set - authenticated - is set to true", async () => {
  const { token } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "user@keel.xyz",
      password: "1234",
    },
  });

  const post = await actions
    .withAuthToken(token)
    .createPostSetIsAuthenticated({ title: "temp" });

  expect(post.isAuthenticated).toEqual(true);
});

test("isAuthenticated context set - not authenticated - is set to false", async () => {
  const post = await actions.createPostSetIsAuthenticated({
    title: "temp",
  });

  expect(post.isAuthenticated).toEqual(false);
});

// todo:  permission test against null.  Requires this fix:  https://linear.app/keel/issue/DEV-195/permissions-support-null-operand-with-identity-type

// todo:  permission test against another identity field.  Requires this fix: https://linear.app/keel/issue/DEV-196/permissions-support-identity-type-operand-with-identity-comparison

test("related model identity context permission - correct identity - permission satisfied", async () => {
  const { token } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "user1@keel.xyz",
      password: "1234",
    },
  });

  const post = await actions
    .withAuthToken(token)
    .createPostWithIdentity({ title: "temp" });

  const child = await actions
    .withAuthToken(token)
    .createChild({ post: { id: post.id } });

  const childPosts = await models.childPost.findMany({ postId: post.id });

  expect(child.postId).toEqual(post.id);
  expect(childPosts.length).toEqual(1);
  expect(childPosts[0].id).toEqual(child.id);
});

test("related model identity context permission - incorrect identity - permission not satisfied", async () => {
  const { token: token1 } = await actions.authenticate({
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

  const post = await actions
    .withAuthToken(token1)
    .createPostWithIdentity({ title: "temp" });

  await expect(
    actions.withAuthToken(token2).createChild({ post: { id: post.id } })
  ).toHaveAuthorizationError();

  const childPosts = await models.childPost.findMany({ postId: post.id });
  expect(childPosts.length).toEqual(0);
});

test("set identity on related models - authenticated - fields set correctly", async () => {
  const authResponse = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "user@keel.xyz",
      password: "1234",
    },
  });

  const authedActions = actions.withAuthToken(authResponse.token);

  const post = await authedActions.createPostWithComments({
    title: "temp",
    comments: [{ comment: "comment 1" }, { comment: "comment 2" }],
  });

  const identity = await models.identity.findOne({ email: "user@keel.xyz" });

  const comments = await models.comment.findMany({});

  expect(comments[0].createdById).toEqual(identity!.id);
  expect(comments[0].isActive).toEqual(true);
  expect(comments[1].createdById).toEqual(identity!.id);
  expect(comments[1].isActive).toEqual(true);
});
