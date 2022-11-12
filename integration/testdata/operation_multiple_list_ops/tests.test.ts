import { test, expect, actions, Thing } from "@teamkeel/testing";

test("allows for two list operations on same model", async () => {
  await Thing.create({ something: "123" });

  const { collection: one } = await actions.listOne({});

  const { collection: two } = await actions.listTwo({});

  expect(one).toEqual(two);
});
