import { actions, resetDatabase } from "@teamkeel/testing";
import { beforeEach, expect, test } from "vitest";

beforeEach(resetDatabase);

test("create person with env var name", async () => {
  const person = await actions.createPerson({});

  expect(person.name).toEqual("Pedro");
});
