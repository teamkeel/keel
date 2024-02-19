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