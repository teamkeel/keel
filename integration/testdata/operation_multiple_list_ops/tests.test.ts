import { test, expect, beforeEach } from "vitest";
import { actions, resetDatabase, models } from "@teamkeel/testing";

beforeEach(resetDatabase);

test("allows for two list actions on same model", async () => {
  await models.thing.create({ something: "123" });

  const { results: one } = await actions.listOne({});

  const { results: two } = await actions.listTwo({});

  expect(one).toEqual(two);
});
