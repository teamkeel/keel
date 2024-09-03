import { actions, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

test("schemas dir - create action", async () => {
  const createdPost = await actions.createPost({
    title: "foo",
    subTitle: "abc",
    content: "# Title",
  });

  expect(createdPost.title).toEqual("foo");
  expect(createdPost.content).toEqual("# Title");
});
