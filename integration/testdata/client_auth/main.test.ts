import { actions, models, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach, beforeAll } from "vitest";
import { APIClient } from "./keelClient";

var client: APIClient;

beforeEach(() => {
  client = new APIClient({ baseUrl: process.env.KEEL_TESTING_CLIENT_API_URL! });
});

beforeEach(resetDatabase);

test("not authenticated - not permitted", async () => {
  const response = await client.api.mutations.createPost({ title: "Test" });
  expect(response.error?.type).toEqual("forbidden");
});

test("not authenticated - permitted", async () => {
  await models.post.create({ title: "Test" });

  const response = await client.api.queries.allPosts();
  expect(response.data?.results).toHaveLength(1);
  expect(response.error).toBeUndefined();
});

test("authenticate action - authenticate not successful", async () => {
  await client.api.mutations.authenticate({
    emailPassword: { email: "user@example.com", password: "1234" },
    createIfNotExists: true,
  });
  const response = await client.api.mutations.authenticate({
    emailPassword: { email: "user@example.com", password: "oops" },
    createIfNotExists: true,
  });
  expect(response.error?.type).toEqual("bad_request");
});

test("authenticate action - authenticate successful", async () => {
  const response = await client.api.mutations.authenticate({
    emailPassword: { email: "user@example.com", password: "1234" },
    createIfNotExists: true,
  });
  expect(response.data?.identityCreated).toEqual(true);
  expect(response.error).toBeUndefined();
  expect(client.ctx.isAuthenticated).toBeTruthy();
  expect(client.ctx.token).toEqual(response.data?.token);
});

test("authenticate action - permitted", async () => {
  const authResponse = await client.api.mutations.authenticate({
    emailPassword: { email: "user@example.com", password: "1234" },
    createIfNotExists: true,
  });
  expect(authResponse.data?.identityCreated).toEqual(true);
  expect(authResponse.error).toBeUndefined();
  expect(client.ctx.isAuthenticated).toBeTruthy();

  const createResponse = await client.api.mutations.createPost({
    title: "Test",
  });
  expect(createResponse.data?.title).toEqual("Test");
});
