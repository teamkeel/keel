import { models, actions, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

async function get(path: string) {
  return fetch(process.env.KEEL_TESTING_API_URL + path, {
    method: "GET",
  });
}

async function post(path: string, body: any) {
  return fetch(process.env.KEEL_TESTING_API_URL + path, {
    method: "POST",
    body: JSON.stringify(body),
  });
}

test("GET route", async () => {
  const r = await get("/get/route?foo=bar");
  expect(r.status).toBe(200);
  const body = await r.json();
  expect(body).toStrictEqual({
    foo: "bar",
  });
});

test("POST route", async () => {
  const r = await post("/post/route", {
    foo: "bar",
  });
  expect(r.status).toBe(200);
  const body = await r.json();
  expect(body).toStrictEqual({
    foo: "bar",
    fizz: "buzz",
  });
});

test("raw body access", async () => {
  const r = await fetch(process.env.KEEL_TESTING_API_URL + "/raw/body/route", {
    method: "POST",
    body: `{"foo": "bar"}`,
  });
  expect(r.status).toBe(200);
  const body = await r.json();
  expect(body).toStrictEqual({
    sha1: "bc4919c6adf7168088eaea06e27a5b23f0f9f9da",
  });
});

test("headers", async () => {
  const r = await fetch(process.env.KEEL_TESTING_API_URL + "/headers/route", {
    method: "POST",
    body: `{}`,
    headers: {
      [`X-My-Request-Header`]: "foo",
    },
  });
  expect(r.status).toBe(200);
  expect(r.headers.get("X-My-Response-Header")).toBe("foobar");
});

test("database access", async () => {
  const r = await post("/database/route", {
    name: "Jon",
  });
  expect(r.status).toBe(200);
  const body = await r.json();

  const p = await models.person.findOne({ id: body.id });
  expect(p!.name).toBe("Jon");
});

test("path params", async () => {
  const r = await get("/path/param/route/bar");
  expect(r.status).toBe(200);
  const body = await r.json();
  expect(body).toStrictEqual({
    foo: "bar",
  });
});

test("response status", async () => {
  const r = await fetch(process.env.KEEL_TESTING_API_URL + "/status/route", {
    method: "PUT",
  });
  expect(r.status).toBe(204);
  const body = await r.text();
  expect(body).toBe("");
});
