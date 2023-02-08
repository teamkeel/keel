import { actions } from "@teamkeel/testing";
import { test, expect } from "vitest";

test("make sure api.fetch works", async () => {
  const fetchedThing = await actions.getFetchedThing({});
  expect(fetchedThing).not.toBeNull();
  const body = fetchedThing?.fetchedBody;
  expect(body?.startsWith("<!doctype html")).toBe(true);
});
