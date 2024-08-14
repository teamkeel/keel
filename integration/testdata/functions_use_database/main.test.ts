import { test, expect } from "vitest";
import { actions, models } from "@teamkeel/testing";

test("testing raw kysely", async () => {
  const postCreatedAfter2020 = await models.post.create({
    title: "a title",
    createdAt: new Date(2020, 2, 1),
    updatedAt: new Date(),
  });

  const goodResult = await actions.getPostButOnlyIfItsAfter2020({
    id: postCreatedAfter2020.id,
  });
  
  expect(goodResult?.id).toEqual(postCreatedAfter2020.id);

  const postCreatedBefore2020 = await models.post.create({
    title: "a title",
    createdAt: new Date(2019, 2, 1),
    updatedAt: new Date(),
  });

  // because the post was created before 2020 and our custom function adds an additional
  // sql constraint to ensure any post found in the db by id was also created after 1/1/2020
  // we expect the return value of the function to be nothing so a no result error will be thrown
  await expect(
    actions.getPostButOnlyIfItsAfter2020({ id: postCreatedBefore2020.id })
  ).rejects.toEqual({
    code: "ERR_RECORD_NOT_FOUND",
    message: "record not found",
  });
});
