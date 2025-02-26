import { models, actions, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

test("GET route", async () => {
  const r = await fetch(
    process.env.KEEL_TESTING_API_URL + "/my/route?someParam=foo",
    {
      method: "GET",
    }
  );
  const body = await r.text();
  expect(body).toBe("query someParam = foo");
});
