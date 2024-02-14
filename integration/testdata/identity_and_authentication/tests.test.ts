import { test, expect, beforeEach } from "vitest";
import { actions, models, resetDatabase } from "@teamkeel/testing";


beforeEach(resetDatabase);

test("create identity", async () => {
  const response = await fetch(process.env.KEEL_TESTING_AUTH_API_URL + "/token", {
    method: "POST",
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      "grant_type": "password",
      "username": "user@keel.xyz",
      "password": "1234"
    })
  });
  expect(response.status).toEqual(200);

  const identity = await models.identity.findOne({ email:  "user@keel.xyz", issuer: "https://keel.so" });
  expect(identity).not.toBeNull();
});

test("authenticate - invalid email - respond with invalid email address error", async () => {
  const response = await fetch(process.env.KEEL_TESTING_AUTH_API_URL + "/token", {
    method: "POST",
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      "grant_type": "password",
      "username": "user",
      "password": "1234"
    })
  });
  expect(response.status).toEqual(400);

  const body = await response.json();

  expect(body).toEqual({
    error: "invalid_request",
    error_description: "invalid email address",
  });
});

test("authenticate - empty password - respond with password cannot be empty error", async () => {
  const response = await fetch(process.env.KEEL_TESTING_AUTH_API_URL + "/token", {
    method: "POST",
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      "grant_type": "password",
      "username": "user@keel.xyz",
      "password": ""
    })
  });
  expect(response.status).toEqual(400);

  const body = await response.json();

  expect(body).toEqual({
    error: "invalid_request",
    error_description: "the identity's password in the 'password' field is required",
  });
});



test("authenticate - existing identity - authenticated", async () => {
  const response1 = await fetch(process.env.KEEL_TESTING_AUTH_API_URL + "/token", {
    method: "POST",
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      "grant_type": "password",
      "username": "user@keel.xyz",
      "password": "1234"
    })
  });
  expect(response1.status).toEqual(200);

  const response2 = await fetch(process.env.KEEL_TESTING_AUTH_API_URL + "/token", {
    method: "POST",
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      "grant_type": "password",
      "username": "user@keel.xyz",
      "password": "1234"
    })
  });
  expect(response2.status).toEqual(200);


  const identities = await models.identity.findMany();
  expect(identities).toHaveLength(1);
});


test("authenticate - incorrect credentials with existing identity - not authenticated", async () => {
  const response1 = await fetch(process.env.KEEL_TESTING_AUTH_API_URL + "/token", {
    method: "POST",
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      "grant_type": "password",
      "username": "user@keel.xyz",
      "password": "1234"
    })
  });
  expect(response1.status).toEqual(200);

  const response2 = await fetch(process.env.KEEL_TESTING_AUTH_API_URL + "/token", {
    method: "POST",
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      "grant_type": "password",
      "username": "user@keel.xyz",
      "password": "wrong"
    })
  });
  expect(response2.status).toEqual(401);

  const body = await response2.json();

  expect(body).toEqual({
    error: "invalid_client",
    error_description: "possible causes may be that the identity does not exist or the credentials are incorrect",
  });

  const identities = await models.identity.findMany();
  expect(identities).toHaveLength(1);
});


test("withAuthToken - invalid token - authentication failed", async () => {
  await expect(
    actions.withAuthToken("invalid").createPostWithIdentity({ title: "temp" })
  ).toHaveAuthenticationError();

  await expect(
    actions
      .withAuthToken("invalid")
      .getPostRequiresAuthentication({ id: "temp" })
  ).toHaveAuthenticationError();

  await expect(
    actions
      .withAuthToken("invalid")
      .getPostRequiresNoAuthentication({ id: "temp" })
  ).toHaveAuthenticationError();

  await expect(
    actions.withAuthToken("invalid").getPostPublic({ id: "temp" })
  ).toHaveAuthenticationError();
});

test("withAuthToken - identity does not exist - authentication failed", async () => {
  const response = await fetch(process.env.KEEL_TESTING_AUTH_API_URL + "/token", {
    method: "POST",
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      "grant_type": "password",
      "username": "user@keel.xyz",
      "password": "1234"
    })
  });
  expect(response.status).toEqual(200);

  const body = await response.json();
  const token = body.access_token;

  await models.identity.delete({
    email: "user@keel.xyz",
    issuer: "https://keel.so",
  });

  await expect(
    actions.withAuthToken(token).createPostWithIdentity({ title: "temp" })
  ).toHaveAuthenticationError();

  await expect(
    actions.withAuthToken("invalid").createPostWithIdentity({ title: "temp" })
  ).toHaveAuthenticationError();

  await expect(
    actions
      .withAuthToken("invalid")
      .getPostRequiresAuthentication({ id: "temp" })
  ).toHaveAuthenticationError();

  await expect(
    actions
      .withAuthToken("invalid")
      .getPostRequiresNoAuthentication({ id: "temp" })
  ).toHaveAuthenticationError();

  await expect(
    actions.withAuthToken("invalid").getPostPublic({ id: "temp" })
  ).toHaveAuthenticationError();
});


test("identity context permission - correct identity - permission satisfied", async () => {
  const response = await fetch(process.env.KEEL_TESTING_AUTH_API_URL + "/token", {
    method: "POST",
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      "grant_type": "password",
      "username": "user@keel.xyz",
      "password": "1234"
    })
  });
  expect(response.status).toEqual(200);

  const body = await response.json();
  const token = body.access_token;

  const authedActions = actions.withAuthToken(token);

  const post = await authedActions.createPostWithIdentity({ title: "temp" });

  await expect(
    authedActions.getPostRequiresIdentity({ id: post.id })
  ).resolves.toEqual(post);
});

test("identity context permission - incorrect identity - permission not satisfied", async () => {
  const response1 = await fetch(process.env.KEEL_TESTING_AUTH_API_URL + "/token", {
    method: "POST",
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      "grant_type": "password",
      "username": "user1@keel.xyz",
      "password": "1234"
    })
  });
  expect(response1.status).toEqual(200);
  const body1 = await response1.json();
  const token1 = body1.access_token;

  const response2 = await fetch(process.env.KEEL_TESTING_AUTH_API_URL + "/token", {
    method: "POST",
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      "grant_type": "password",
      "username": "user2@keel.xyz",
      "password": "1234"
    })
  });
  expect(response2.status).toEqual(200);
  const body2 = await response2.json();
  const token2 = body2.access_token;

  const post = await actions
    .withAuthToken(token1)
    .createPostWithIdentity({ title: "temp" });

  await expect(
    actions.withAuthToken(token2).getPostRequiresIdentity({ id: post.id })
  ).toHaveAuthorizationError();
});

test("isAuthenticated context permission - authenticated - permission satisfied", async () => {
  const response = await fetch(process.env.KEEL_TESTING_AUTH_API_URL + "/token", {
    method: "POST",
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      "grant_type": "password",
      "username": "user@keel.xyz",
      "password": "1234"
    })
  });
  expect(response.status).toEqual(200);

  const body = await response.json();
  const token = body.access_token;


  const post = await actions
    .withAuthToken(token)
    .createPostWithIdentity({ title: "temp" });

  await expect(
    actions.withAuthToken(token).getPostRequiresAuthentication({ id: post.id })
  ).resolves.toEqual(post);
});

test("isAuthenticated context permission - not authenticated - permission not satisfied", async () => {
  const response = await fetch(process.env.KEEL_TESTING_AUTH_API_URL + "/token", {
    method: "POST",
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      "grant_type": "password",
      "username": "user@keel.xyz",
      "password": "1234"
    })
  });
  expect(response.status).toEqual(200);

  const body = await response.json();
  const token = body.access_token;


  const post = await actions
    .withAuthToken(token)
    .createPostWithIdentity({ title: "temp" });

  await expect(
    actions.getPostRequiresAuthentication({ id: post.id })
  ).toHaveAuthorizationError();
});

test("not isAuthenticated context permission - authenticated - permission satisfied", async () => {
  const response = await fetch(process.env.KEEL_TESTING_AUTH_API_URL + "/token", {
    method: "POST",
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      "grant_type": "password",
      "username": "user@keel.xyz",
      "password": "1234"
    })
  });
  expect(response.status).toEqual(200);

  const body = await response.json();
  const token = body.access_token;


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
  const response = await fetch(process.env.KEEL_TESTING_AUTH_API_URL + "/token", {
    method: "POST",
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      "grant_type": "password",
      "username": "user@keel.xyz",
      "password": "1234"
    })
  });
  expect(response.status).toEqual(200);

  const body = await response.json();
  const token = body.access_token;

  const post = await actions
    .withAuthToken(token)
    .createPostWithIdentity({ title: "temp" });

  await expect(
    actions.getPostRequiresNoAuthentication({ id: post.id })
  ).resolves.toEqual(post);
});

test("isAuthenticated context set - authenticated - is set to true", async () => {
  const response = await fetch(process.env.KEEL_TESTING_AUTH_API_URL + "/token", {
    method: "POST",
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      "grant_type": "password",
      "username": "user@keel.xyz",
      "password": "1234"
    })
  });
  expect(response.status).toEqual(200);

  const body = await response.json();
  const token = body.access_token;


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

test("related model identity context permission - correct identity - permission satisfied", async () => {
  const response = await fetch(process.env.KEEL_TESTING_AUTH_API_URL + "/token", {
    method: "POST",
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      "grant_type": "password",
      "username": "user@keel.xyz",
      "password": "1234"
    })
  });
  expect(response.status).toEqual(200);

  const body = await response.json();
  const token = body.access_token;


  const post = await actions
    .withAuthToken(token)
    .createPostWithIdentity({ title: "temp" });

  const child = await actions
    .withAuthToken(token)
    .createChild({ post: { id: post.id } });

  const childPosts = await models.childPost.findMany({
    where: { postId: post.id },
  });

  expect(child.postId).toEqual(post.id);
  expect(childPosts.length).toEqual(1);
  expect(childPosts[0].id).toEqual(child.id);
});

test("related model identity context permission - incorrect identity - permission not satisfied", async () => {
  const response1 = await fetch(process.env.KEEL_TESTING_AUTH_API_URL + "/token", {
    method: "POST",
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      "grant_type": "password",
      "username": "user1@keel.xyz",
      "password": "1234"
    })
  });
  expect(response1.status).toEqual(200);
  const body1 = await response1.json();
  const token1 = body1.access_token;

  const response2 = await fetch(process.env.KEEL_TESTING_AUTH_API_URL + "/token", {
    method: "POST",
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      "grant_type": "password",
      "username": "user2@keel.xyz",
      "password": "1234"
    })
  });
  expect(response2.status).toEqual(200);
  const body2 = await response2.json();
  const token2 = body2.access_token;

  const post = await actions
    .withAuthToken(token1)
    .createPostWithIdentity({ title: "temp" });

  await expect(
    actions.withAuthToken(token2).createChild({ post: { id: post.id } })
  ).toHaveAuthorizationError();

  const childPosts = await models.childPost.findMany({
    where: { postId: post.id },
  });
  expect(childPosts.length).toEqual(0);
});

test("request reset password - invalid email - respond with invalid email address error", async () => {
  await expect(
    actions.requestPasswordReset({
      email: "user",
      redirectUrl: "https://mydomain.com",
    })
  ).rejects.toEqual({
    code: "ERR_INVALID_INPUT",
    message: "invalid email address",
  });
});

test("request reset password - invalid redirectUrl - respond with invalid redirectUrl error", async () => {
  await expect(
    actions.requestPasswordReset({
      email: "user@keel.xyz",
      redirectUrl: "mydomain",
    })
  ).rejects.toEqual({
    code: "ERR_INVALID_INPUT",
    message: "invalid redirect URL",
  });
});

test("request reset password - unknown email - successful request", async () => {
  await models.identity.create({
    email: "user@keel.xyz",
    password: "123",
  });

  await expect(
    actions.requestPasswordReset({
      email: "another-user@keel.xyz",
      redirectUrl: "https://mydomain.com",
    })
  ).not.toHaveError({});
});

// This test will break if we use a private key in the test runtime.
test("reset password - invalid token - cannot be parsed", async () => {
  const identity = await models.identity.create({
    id: "2OrbbxUb8syZzlDz0v5ofunO1vi",
    email: "user@keel.xyz",
    password: "123",
  });

  await expect(
    actions.resetPassword({
      token: "invalid",
      password: "abc",
    })
  ).rejects.toEqual({
    code: "ERR_INVALID_INPUT",
    message: "cannot be parsed or verified as a valid JWT",
  });
});

// This test will break if we use a private key in the test runtime.
test("reset password - missing aud claim - cannot be parsed error", async () => {
  const identity = await models.identity.create({
    id: "2OrbbxUb8syZzlDz0v5ofunO1vi",
    email: "user@keel.xyz",
    password: "123",
  });

  // {
  //   "typ": "JWT",
  //   "alg": "none"
  // }
  // {
  //   "sub": "2OrbbxUb8syZzlDz0v5ofunO1vi",
  //   "iat": 1682323697,
  //   "exp": 1893459661
  // }
  const resetToken =
    "eyJ0eXAiOiJKV1QiLCJhbGciOiJub25lIn0.eyJzdWIiOiIyT3JiYnhVYjhzeVp6bER6MHY1b2Z1bk8xdmkiLCJpYXQiOjE2ODIzMjM2OTcsImV4cCI6MTg5MzQ1OTY2MX0.";

  await expect(
    actions.resetPassword({
      token: resetToken,
      password: "abc",
    })
  ).rejects.toEqual({
    code: "ERR_INVALID_INPUT",
    message: "cannot be parsed or verified as a valid JWT",
  });
});

// This test will break if we use a private key in the test runtime.
// test("reset password - valid token - password is reset", async () => {
//   const identity = await models.identity.create({
//     id: "2OrbbxUb8syZzlDz0v5ofunO1vi",
//     email: "user@keel.xyz",
//     password: "123",
//   });

//   // {
//   //   "typ": "JWT",
//   //   "alg": "none"
//   // }
//   // {
//   //   "sub": "2OrbbxUb8syZzlDz0v5ofunO1vi",
//   //   "iat": 1682323697,
//   //   "exp": 1893459661,
//   //   "aud": "password-reset"
//   // }
//   const resetToken =
//     "eyJ0eXAiOiJKV1QiLCJhbGciOiJub25lIn0.eyJzdWIiOiIyT3JiYnhVYjhzeVp6bER6MHY1b2Z1bk8xdmkiLCJpYXQiOjE2ODIzMjM2OTcsImV4cCI6MTg5MzQ1OTY2MSwiYXVkIjoicGFzc3dvcmQtcmVzZXQifQ.";

//   await expect(
//     actions.resetPassword({
//       token: resetToken,
//       password: "abc",
//     })
//   ).not.toHaveError({});

//   await expect(
//     actions.authenticate({
//       createIfNotExists: false,
//       emailPassword: {
//         email: "user@keel.xyz",
//         password: "123",
//       },
//     })
//   ).rejects.toEqual({
//     code: "ERR_INVALID_INPUT",
//     message: "failed to authenticate",
//   });

//   const { token } = await actions.authenticate({
//     createIfNotExists: false,
//     emailPassword: {
//       email: "user@keel.xyz",
//       password: "abc",
//     },
//   });

//   expect(token).not.toBeNull();
// });

// This test will break if we use a private key in the test runtime.
// test("reset password - valid token with aud as array - password is reset", async () => {
//   const identity = await models.identity.create({
//     id: "2OrbbxUb8syZzlDz0v5ofunO1vi",
//     email: "user@keel.xyz",
//     password: "123",
//   });

//   // {
//   //   "typ": "JWT",
//   //   "alg": "none"
//   // }
//   // {
//   //   "sub": "2OrbbxUb8syZzlDz0v5ofunO1vi",
//   //   "iat": 1682323697,
//   //   "exp": 1893459661,
//   //   "aud": ["password-reset"]
//   // }
//   const resetToken =
//     "eyJ0eXAiOiJKV1QiLCJhbGciOiJub25lIn0.eyJzdWIiOiIyT3JiYnhVYjhzeVp6bER6MHY1b2Z1bk8xdmkiLCJpYXQiOjE2ODIzMjM2OTcsImV4cCI6MTg5MzQ1OTY2MSwiYXVkIjpbInBhc3N3b3JkLXJlc2V0Il19.";

//   await expect(
//     actions.resetPassword({
//       token: resetToken,
//       password: "abc",
//     })
//   ).not.toHaveError({});

//   await expect(
//     actions.authenticate({
//       createIfNotExists: false,
//       emailPassword: {
//         email: "user@keel.xyz",
//         password: "123",
//       },
//     })
//   ).rejects.toEqual({
//     code: "ERR_INVALID_INPUT",
//     message: "failed to authenticate",
//   });

//   const { token } = await actions.authenticate({
//     createIfNotExists: false,
//     emailPassword: {
//       email: "user@keel.xyz",
//       password: "abc",
//     },
//   });

//   expect(token).not.toBeNull();
// });

test("create and authenticate - email exists for another issuer - success", async () => {
  await models.identity.create({
    email: "keel@keelson.so",
    issuer: "https://auth.staging.keel.xyz/",
    externalId: "google-oauth2|117415937240512761581",
  });

  const response = await fetch(process.env.KEEL_TESTING_AUTH_API_URL + "/token", {
    method: "POST",
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      "grant_type": "password",
      "username": "keel@keelson.so",
      "password": "1234"
    })
  });
  expect(response.status).toEqual(200);

  const identities = await models.identity.findMany();
  expect(identities).toHaveLength(2);
});

test("identity with custom non-ksuid id", async () => {
  const johnDoe = await models.identity.create({
    id: "not-a-ksuid",
    email: "john@example.com",
    issuer: "https://keel.so"
  });

  const post = await models.post.create({
    title: "example post",
    identityId: johnDoe.id,
  });

  const fetchedRecord = await actions
    .withIdentity(johnDoe)
    .getPostRequiresIdentity({ id: post.id });
  expect(fetchedRecord).toEqual(post);

  const fetchedIdentity = await models.identity.findOne({ id: "not-a-ksuid" });
  expect(fetchedIdentity).toEqual(johnDoe);
});
