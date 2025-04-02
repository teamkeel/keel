import { actions, resetDatabase, models, flows } from "@teamkeel/testing";
import { beforeEach, expect, test } from "vitest";

beforeEach(resetDatabase);

test("flows - basic execution", async () => {
  await flows.myFlow({ name: "Keelson", age: 25 });

  const things = await models.thing.findMany();
  expect(things.length).toBe(1);
  expect(things[0].name).toBe("Keelson");
});
