import { actions, models, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach, beforeAll } from "vitest";
import { APIClient } from "./keelClient";

const baseUrl = process.env.KEEL_TESTING_CLIENT_API_URL!;

beforeEach(resetDatabase);

test("authenticateWithPassword", async () => {
  const store = new TokenStore();
  const client = new APIClient({ baseUrl }, store.getTokens, store.setTokens);

  await client.auth.authenticateWithPassword("user@example.com", "1234");
  expect(await client.auth.isAuthenticated()).toBeTruthy();
  expect(await client.auth.expiresAt()).not.toBeNull();
  expect((await client.auth.expiresAt()!) > new Date()).toBeTruthy();
  expect(store.accessToken).not.toBeNull();
  expect(store.refreshToken).not.toBeNull();

  await client.auth.authenticateWithPassword("user@example.com", "oops");
  expect(await client.auth.isAuthenticated()).not.toBeTruthy();
  expect(await client.auth.expiresAt()).toBeNull();
  expect(store.accessToken).toBeNull();
  expect(store.refreshToken).toBeNull();

  await models.post.create({ title: "Test" });
  const response1 = await client.api.queries.allPosts();
  expect(response1.error!.type).toEqual("forbidden");

  await client.auth.authenticateWithPassword("user@example.com", "1234");
  expect(await client.auth.isAuthenticated()).toBeTruthy();
  expect(await client.auth.expiresAt()).not.toBeNull();
  expect((await client.auth.expiresAt()!) > new Date()).toBeTruthy();
  expect(store.accessToken).not.toBeNull();
  expect(store.refreshToken).not.toBeNull();

  const response2 = await client.api.queries.allPosts();
  expect(response2.data?.results).toHaveLength(1);
});

test("valid access token", async () => {
  const store = new TokenStore();
  const client = new APIClient({ baseUrl }, store.getTokens, store.setTokens);

  await client.auth.authenticateWithPassword("user@example.com", "1234");
  expect(await client.auth.isAuthenticated()).toBeTruthy();

  expect(store.accessToken).not.toBeNull();
  expect(store.refreshToken).not.toEqual("");
});

test("valid refresh token", async () => {
  const store = new TokenStore();
  const client = new APIClient({ baseUrl }, store.getTokens, store.setTokens);

  await client.auth.authenticateWithPassword("user@example.com", "1234");
  expect(await client.auth.isAuthenticated()).toBeTruthy();

  expect(store.refreshToken).not.toBeNull();
  expect(store.refreshToken).not.toEqual("");
});

test("refreshing successfully", async () => {
  const store = new TokenStore();
  const client = new APIClient({ baseUrl }, store.getTokens, store.setTokens);

  await client.auth.authenticateWithPassword("user@example.com", "1234");
  expect(await client.auth.isAuthenticated()).toBeTruthy();

  const accessToken = store.accessToken;
  const refreshToken = store.refreshToken;

  expect(accessToken).not.toBeNull();
  expect(refreshToken).not.toBeNull();

  const expiry1 = client.auth.expiresAt();
  expect(expiry1).not.toBeNull();

  await delay(1000);
  const refreshed = await client.auth.refresh();
  expect(refreshed).toBeTruthy();

  expect(store.accessToken).not.toBeNull();
  expect(store.refreshToken).not.toBeNull();
  expect(store.accessToken).not.toEqual(accessToken);
  expect(store.refreshToken).not.toEqual(refreshToken);

  const expiry2 = client.auth.expiresAt();
  expect(expiry1?.getTime()).lessThan(expiry2!.getTime());
});

test("logout successfully", async () => {
  const store = new TokenStore();
  const client = new APIClient({ baseUrl }, store.getTokens, store.setTokens);

  await client.auth.authenticateWithPassword("user@example.com", "1234");
  expect(await client.auth.isAuthenticated()).toBeTruthy();

  await client.auth.logout();

  expect(store.accessToken).toBeNull();
  expect(store.refreshToken).toBeNull();

  expect(await client.auth.expiresAt()).toBeNull();
  expect(await client.auth.isAuthenticated()).not.toBeTruthy();
});

test("logout revokes refresh token successfully", async () => {
  const store = new TokenStore();
  const client = new APIClient({ baseUrl }, store.getTokens, store.setTokens);

  await client.auth.authenticateWithPassword("user@example.com", "1234");
  expect(await client.auth.isAuthenticated()).toBeTruthy();

  await client.auth.logout();

  const accessToken = store.accessToken;
  const refreshToken = store.refreshToken;

  expect(store.accessToken).toBeNull();
  expect(store.refreshToken).toBeNull();

  expect(await client.auth.expiresAt()).toBeNull();
  expect(await client.auth.isAuthenticated()).not.toBeTruthy();

  store.setTokens(accessToken, refreshToken);

  const refresh = await client.auth.refresh();
  expect(refresh).not.toBeTruthy();
  expect(await client.auth.isAuthenticated()).not.toBeTruthy();

  expect(store.accessToken).toBeNull();
  expect(store.refreshToken).toBeNull();
});

test("authentication flow with default token store", async () => {
  const client = new APIClient({ baseUrl });

  await client.auth.authenticateWithPassword("user@example.com", "1234");
  expect(await client.auth.isAuthenticated()).toBeTruthy();
  expect(await client.auth.expiresAt()).not.toBeNull();
  expect((await client.auth.expiresAt()!) > new Date()).toBeTruthy();

  await client.auth.authenticateWithPassword("user@example.com", "oops");
  expect(await client.auth.isAuthenticated()).not.toBeTruthy();
  expect(await client.auth.expiresAt()).toBeNull();

  await client.auth.authenticateWithPassword("user@example.com", "1234");
  expect(await client.auth.isAuthenticated()).toBeTruthy();
  expect(await client.auth.expiresAt()).not.toBeNull();
  expect((await client.auth.expiresAt()!) > new Date()).toBeTruthy();

  const expiry1 = client.auth.expiresAt();

  await delay(1000);
  const refreshed = await client.auth.refresh();
  expect(refreshed).toBeTruthy();
  expect(await client.auth.isAuthenticated()).toBeTruthy();

  const expiry2 = client.auth.expiresAt();
  expect(expiry1?.getTime()).lessThan(expiry2!.getTime());

  await client.auth.logout();
  expect(await client.auth.isAuthenticated()).not.toBeTruthy();

  const refreshed2 = await client.auth.refresh();
  expect(refreshed2).not.toBeTruthy();
  expect(await client.auth.isAuthenticated()).not.toBeTruthy();
});

class TokenStore {
  public accessToken: string | null = null;
  public refreshToken: string | null = null;

  getTokens = () => {
    return {
      accessToken: this.accessToken,
      refreshToken: this.refreshToken,
    };
  };

  setTokens = (accessToken: string | null, refreshToken: string | null) => {
    this.accessToken = accessToken;
    this.refreshToken = refreshToken;
  };
}

function delay(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
