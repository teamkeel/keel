import { test, expect } from "vitest";

test("headers", async () => {
  const response = await fetch(
    process.env.KEEL_TESTING_ACTIONS_API_URL + "/withHeaders",
    {
      body: "{}",
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-MyRequestHeader": "my-header-value",
      },
    }
  );

  expect(response.status).toEqual(200);
  expect(response.headers.get("X-MyResponseHeader")).toEqual("my-header-value");
});

test("status", async () => {
  const response = await fetch(
    process.env.KEEL_TESTING_ACTIONS_API_URL + "/withStatus",
    {
      body: "{}",
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      redirect: "manual",
    }
  );

  expect(response.status).toEqual(301);
  expect(response.headers.get("Location")).toEqual("https://some.url");
});

test("query params", async () => {
  const response = await fetch(
    process.env.KEEL_TESTING_ACTIONS_API_URL +
      "/withQueryParams?a=1&b=foo&c=true"
  );

  const body = await response.json();

  expect(body).toEqual({
    a: "1",
    b: "foo",
    c: "true",
  });
});



test("x-www-form-urlencoded", async () => {
  const response = await fetch(
    process.env.KEEL_TESTING_ACTIONS_API_URL +"/withForm",
    {
      method: "POST",
      headers: {
        "Content-Type": "application/x-www-form-urlencoded",
      },
      body: new URLSearchParams({
        'a': '1',
        'b': 'foo',
        'c': 'true'
      })
    }
  );

  const body = await response.json();

  expect(body).toEqual({
    a: "1",
    b: "foo",
    c: "true",
  });
});
