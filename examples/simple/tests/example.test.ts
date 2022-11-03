import { test, expect, actions, Post } from "@teamkeel/testing";

test("create action", async () => {
  const { object: createdPost } = await actions.createPost({
    title: "apple",
    subTitle: "abc",
  });

  expect(createdPost.title).toEqual("apple");
});

test("list action", async () => {
  await Post.create({ title: "apple" });
  await Post.create({ title: "apple" });

  const { collection } = await actions.listPosts({
    title: { equals: "apple" },
  });

  expect(collection.length).toEqual(2);
});
