import { actions, models, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach, beforeAll } from "vitest";
import { APIClient } from "./keelClient";

var client: APIClient;

beforeEach(() => {
  client = new APIClient({ baseUrl: process.env.KEEL_TESTING_CLIENT_API_URL! });
});

beforeEach(resetDatabase);

test("authentication - forbidden", async () => {
  const res = await client.auth.authenticateWithPassword({
    email: "user1@example.com",
    password: "1234",
  });
  expect(res.data?.identityCreated).toEqual(true);
  expect(res.error).toBeUndefined();

  const isAuthed = await client.auth.isAuthenticated();
  expect(isAuthed.data).toBeTruthy();
  expect(isAuthed.error).toBeUndefined();

  const response1 = await client.api.mutations.createPost({ title: "Test" });
  expect(response1.data).not.toBeNull();

  const response2 = await client.api.queries.getPost({
    id: response1.data!.id,
  });
  expect(response2.data?.id).toEqual(response1.data!.id);

  await client.auth.logout();
  const isAuthed2 = await client.auth.isAuthenticated();
  expect(isAuthed2.data).not.toBeTruthy();
  expect(isAuthed2.error).toBeUndefined();

  const response3 = await client.api.queries.getPost({
    id: response1.data!.id,
  });
  expect(response3.error?.type).toEqual("forbidden");

  const res2 = await client.auth.authenticateWithPassword({
    email: "user2@example.com",
    password: "1234",
  });
  expect(res2.data?.identityCreated).toEqual(true);
  expect(res2.error).toBeUndefined();

  const isAuthed3 = await client.auth.isAuthenticated();
  expect(isAuthed3.data).toBeTruthy();
  expect(isAuthed3.error).toBeUndefined();

  const response4 = await client.api.queries.getPost({
    id: response1.data!.id,
  });
  expect(response4.error?.type).toEqual("forbidden");

  const res3 = await client.auth.authenticateWithPassword({
    email: "user2@example.com",
    password: "1234",
  });
  expect(res3.data?.identityCreated).toEqual(false);
  expect(res3.error).toBeUndefined();

  const isAuthed4 = await client.auth.isAuthenticated();
  expect(isAuthed4.data).toBeTruthy();
  expect(isAuthed4.error).toBeUndefined();
});

test("authentication - not authenticated and no permissions", async () => {
  await models.post.create({ title: "Test" });

  const response = await client.api.queries.allPosts();
  expect(response.data?.results).toHaveLength(1);
  expect(response.error).toBeUndefined();
});

test("authentication - get providers", async () => {
  const provs = await client.auth.providers();
  console.log(provs);

  expect(provs.data?.[0]).toEqual({
    name: "myOidcProvider",
    type: "oidc",
    authorizeUrl: process.env.KEEL_TESTING_AUTH_API_URL + "/authorize/myoidcprovider",
    callbackUrl: process.env.KEEL_TESTING_AUTH_API_URL + "/callback/myoidcprovider",
  });
  
  expect(provs.data?.[1]).toEqual({
    name: "googleProvider",
    type: "google",
    authorizeUrl: process.env.KEEL_TESTING_AUTH_API_URL + "/authorize/googleprovider",
    callbackUrl: process.env.KEEL_TESTING_AUTH_API_URL + "/callback/googleprovider",
  });

  expect(provs.data?.[2]).toEqual({
    name: "slackProvider",
    type: "slack",
    authorizeUrl: process.env.KEEL_TESTING_AUTH_API_URL + "/authorize/slackprovider",
    callbackUrl: process.env.KEEL_TESTING_AUTH_API_URL + "/callback/slackprovider",
  });
  
  expect(provs.data?.[3]).toEqual({
    name: "facebookProvider",
    type: "facebook",
    authorizeUrl: process.env.KEEL_TESTING_AUTH_API_URL + "/authorize/facebookprovider",
    callbackUrl: process.env.KEEL_TESTING_AUTH_API_URL + "/callback/facebookprovider",
  });

  expect(provs.data?.[4]).toEqual({
    name: "Gitlab_Provider",
    type: "gitlab",
    authorizeUrl: process.env.KEEL_TESTING_AUTH_API_URL + "/authorize/gitlab_provider",
    callbackUrl: process.env.KEEL_TESTING_AUTH_API_URL + "/callback/gitlab_provider",
  });

  expect(provs.data?.length).toEqual(5);
});