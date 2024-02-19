import { actions, models, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach, beforeAll } from "vitest";
import { APIClient } from "./keelClient";

var client: APIClient;

beforeEach(() => {
  client = new APIClient({ baseUrl: process.env.KEEL_TESTING_CLIENT_API_URL! });
});

beforeEach(resetDatabase);

test("authentication - forbidden", async () => {
  await client.auth.authenticateWithPassword("user1@example.com", "1234");
  expect(await client.auth.isAuthenticated()).toBeTruthy();
  const response1 = await client.api.mutations.createPost({ title: "Test" });
  expect(response1.data).not.toBeNull();

  const response2 = await client.api.queries.getPost({
    id: response1.data!.id,
  });
  expect(response2.data?.id).toEqual(response1.data!.id);

  await client.auth.logout();
  expect(await client.auth.isAuthenticated()).not.toBeTruthy();

  const response3 = await client.api.queries.getPost({
    id: response1.data!.id,
  });
  expect(response3.error?.type).toEqual("forbidden");

  await client.auth.authenticateWithPassword("user2@example.com", "1234");
  expect(await client.auth.isAuthenticated()).toBeTruthy();

  const response4 = await client.api.queries.getPost({
    id: response1.data!.id,
  });
  expect(response4.error?.type).toEqual("forbidden");
});

test("authentication - not authenticated and no permissions", async () => {
  await models.post.create({ title: "Test" });

  const response = await client.api.queries.allPosts();
  expect(response.data?.results).toHaveLength(1);
  expect(response.error).toBeUndefined();
});
