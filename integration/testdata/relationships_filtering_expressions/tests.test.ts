import { actions, models, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

test("get action where expressions with M:1 relations - all models active - model returned", async () => {
  const publisher = await models.publisher.create({
    orgName: "Keel Org",
    isActive: true,
  });
  const author = await models.author.create({
    name: "Keelson",
    thePublisherId: publisher.id,
    isActive: true,
  });
  const firstPost = await models.post.create({
    title: "My First Post",
    theAuthorId: author.id,
    isActive: true,
  });

  const post = await actions.getActivePost({ id: firstPost.id });

  expect(post!.id).toEqual(firstPost.id);
});

test("get action where expressions with M:1 relations - post model not active - no records found", async () => {
  const publisher = await models.publisher.create({
    orgName: "Keel Org",
    isActive: true,
  });
  const author = await models.author.create({
    name: "Keelson",
    thePublisherId: publisher.id,
    isActive: true,
  });
  const firstPost = await models.post.create({
    title: "My First Post",
    theAuthorId: author.id,
    isActive: false,
  });

  const p = await actions.getActivePost({ id: firstPost.id });
  expect(p).toEqual(null);
});

test("get action where expressions with M:1 relations - nested author model not active - no records found", async () => {
  const publisher = await models.publisher.create({
    orgName: "Keel Org",
    isActive: true,
  });
  const author = await models.author.create({
    name: "Keelson",
    thePublisherId: publisher.id,
    isActive: false,
  });
  const firstPost = await models.post.create({
    title: "My First Post",
    theAuthorId: author.id,
    isActive: true,
  });

  const p = await actions.getActivePost({ id: firstPost.id });
  expect(p).toEqual(null);
});

test("get action where expressions with M:1 relations - nested nested publisher model not active - no records found", async () => {
  const publisher = await models.publisher.create({
    orgName: "Keel Org",
    isActive: false,
  });
  const author = await models.author.create({
    name: "Keelson",
    thePublisherId: publisher.id,
    isActive: true,
  });
  const firstPost = await models.post.create({
    title: "My First Post",
    theAuthorId: author.id,
    isActive: true,
  });

  const p = await actions.getActivePost({ id: firstPost.id });
  expect(p).toEqual(null);
});

test("get action where expressions with 1:M relations - all models active - publisher returned", async () => {
  const publisherKeel = await models.publisher.create({
    orgName: "Keel Org",
    isActive: true,
  });
  const author1 = await models.author.create({
    name: "Keelson",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const publisher = await actions.getActivePublisherWithActivePosts({
    id: publisherKeel.id,
  });

  expect(publisher!.id).toEqual(publisherKeel.id);
});

test("get action where expressions with 1:M relations - publisher not active - no publisher found", async () => {
  const publisherKeel = await models.publisher.create({
    orgName: "Keel Org",
    isActive: false,
  });
  const author1 = await models.author.create({
    name: "Keelson",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const p = await actions.getActivePublisherWithActivePosts({
    id: publisherKeel.id,
  });
  expect(p).toEqual(null);
});

test("get action where expressions with 1:M relations - one author active - publisher returned", async () => {
  const publisherKeel = await models.publisher.create({
    orgName: "Keel Org",
    isActive: true,
  });
  const author1 = await models.author.create({
    name: "Keelson",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    thePublisherId: publisherKeel.id,
    isActive: false,
  });
  await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const publisher = await actions.getActivePublisherWithActivePosts({
    id: publisherKeel.id,
  });

  expect(publisher!.id).toEqual(publisherKeel.id);
});

test("get action where expressions with 1:M relations - active author with inactive posts and inactive autor with active posts - no publisher found", async () => {
  const publisherKeel = await models.publisher.create({
    orgName: "Keel Org",
    isActive: true,
  });
  const author1 = await models.author.create({
    name: "Keelson",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    thePublisherId: publisherKeel.id,
    isActive: false,
  });
  await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: false,
  });
  await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: false,
  });
  await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const p = await actions.getActivePublisherWithActivePosts({
    id: publisherKeel.id,
  });
  expect(p).toEqual(null);
});

test("get action where expressions with 1:M relations - no active posts  - publisher returned", async () => {
  const publisherKeel = await models.publisher.create({
    orgName: "Keel Org",
    isActive: true,
  });
  const author1 = await models.author.create({
    name: "Keelson",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: false,
  });
  await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: false,
  });
  await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: false,
  });

  const p = await actions.getActivePublisherWithActivePosts({
    id: publisherKeel.id,
  });
  expect(p).toEqual(null);
});

test("list action where expressions with M:1 relations - all models active - all models returned", async () => {
  const publisherKeel = await models.publisher.create({
    orgName: "Keel Org",
    isActive: true,
  });
  const publisherWeave = await models.publisher.create({
    orgName: "Weave Org",
    isActive: true,
  });
  const author1 = await models.author.create({
    name: "Keelson",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    thePublisherId: publisherWeave.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const { results: posts } = await actions.listActivePosts({});

  expect(posts.length).toEqual(3);
});

test("list action where expressions with M:1 relations - Keel org not active - Weave models returned", async () => {
  const publisherKeel = await models.publisher.create({
    orgName: "Keel Org",
    isActive: false,
  });
  const publisherWeave = await models.publisher.create({
    orgName: "Weave Org",
    isActive: true,
  });
  const author1 = await models.author.create({
    name: "Keelson",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    thePublisherId: publisherWeave.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const { results: posts } = await actions.listActivePosts({});

  expect(posts.length).toEqual(1);
});

test("list action where expressions with M:1 relations - Keelson author not active - Weaveton models returned", async () => {
  const publisherKeel = await models.publisher.create({
    orgName: "Keel Org",
    isActive: true,
  });
  const publisherWeave = await models.publisher.create({
    orgName: "Weave Org",
    isActive: true,
  });
  const author1 = await models.author.create({
    name: "Keelson",
    thePublisherId: publisherKeel.id,
    isActive: false,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    thePublisherId: publisherWeave.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const { results: posts } = await actions.listActivePosts({});

  expect(posts.length).toEqual(1);
});

test("list action where expressions with M:1 relations - one Keelson post not active - Weaveton models returned", async () => {
  const publisherKeel = await models.publisher.create({
    orgName: "Keel Org",
    isActive: true,
  });
  const publisherWeave = await models.publisher.create({
    orgName: "Weave Org",
    isActive: true,
  });
  const author1 = await models.author.create({
    name: "Keelson",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    thePublisherId: publisherWeave.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: false,
  });
  await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const { results: posts } = await actions.listActivePosts({});

  expect(posts.length).toEqual(2);
});

test("list action where expressions with 1:M relations - all models active - all models returned", async () => {
  const publisherKeel = await models.publisher.create({
    orgName: "Keel Org 2",
    isActive: true,
  });
  const publisherWeave = await models.publisher.create({
    orgName: "Weave Org 2",
    isActive: true,
  });
  const author1 = await models.author.create({
    name: "Keelson",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    thePublisherId: publisherWeave.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const { results: publishers } =
    await actions.listActivePublishersWithActivePosts({});

  expect(publishers.length).toEqual(2);
});

test("list action where expressions with 1:M relations - Keel org not active - only Keel returned", async () => {
  const publisherKeel = await models.publisher.create({
    orgName: "Keel Org",
    isActive: false,
  });
  const publisherWeave = await models.publisher.create({
    orgName: "Weave Org",
    isActive: true,
  });
  const author1 = await models.author.create({
    name: "Keelson",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    thePublisherId: publisherWeave.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const { results: publishers } =
    await actions.listActivePublishersWithActivePosts({});

  expect(publishers.length).toEqual(1);
});

test("list action where expressions with 1:M relations - Keel author not active - Weave org returned", async () => {
  const publisherKeel = await models.publisher.create({
    orgName: "Keel Org",
    isActive: true,
  });
  const publisherWeave = await models.publisher.create({
    orgName: "Weave Org",
    isActive: true,
  });
  const author1 = await models.author.create({
    name: "Keelson",
    thePublisherId: publisherKeel.id,
    isActive: false,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    thePublisherId: publisherWeave.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const { results: publishers } =
    await actions.listActivePublishersWithActivePosts({});

  expect(publishers.length).toEqual(1);
});

test("list action where expressions with 1:M relations - one Keel post not active - all models returned", async () => {
  const publisherKeel = await models.publisher.create({
    orgName: "Keel Org",
    isActive: true,
  });
  const publisherWeave = await models.publisher.create({
    orgName: "Weave Org",
    isActive: true,
  });
  const author1 = await models.author.create({
    name: "Keelson",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    thePublisherId: publisherWeave.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: false,
  });
  await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const { results: publishers } =
    await actions.listActivePublishersWithActivePosts({});

  expect(publishers.length).toEqual(2);
});

test("list action where expressions with 1:M relations - all Keel posts not active - Weave org returned", async () => {
  const publisherKeel = await models.publisher.create({
    orgName: "Keel Org",
    isActive: true,
  });
  const publisherWeave = await models.publisher.create({
    orgName: "Weave Org",
    isActive: true,
  });
  const author1 = await models.author.create({
    name: "Keelson",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    thePublisherId: publisherWeave.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: false,
  });
  await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: false,
  });
  await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const { results: publishers } =
    await actions.listActivePublishersWithActivePosts({});

  expect(publishers.length).toEqual(1);
});

test("get action where expressions with M:1 relations with RHS field operand - all models active - model returned", async () => {
  const publisher = await models.publisher.create({
    orgName: "Keel Org",
    isActive: true,
    booleanValue: true,
  });
  const author = await models.author.create({
    name: "Keelson",
    thePublisherId: publisher.id,
    isActive: true,
  });
  const firstPost = await models.post.create({
    title: "My First Post",
    theAuthorId: author.id,
    isActive: true,
  });

  const post = await actions.getActivePostWithRhsField({
    id: firstPost.id,
  });

  expect(post!.id).toEqual(firstPost.id);
});

test("get action where expressions with M:1 relations with RHS field operand - all models inactive - model returned", async () => {
  const publisher = await models.publisher.create({
    orgName: "Keel Org",
    isActive: false,
    booleanValue: false,
  });
  const author = await models.author.create({
    name: "Keelson",
    thePublisherId: publisher.id,
    isActive: false,
  });
  const firstPost = await models.post.create({
    title: "My First Post",
    theAuthorId: author.id,
    isActive: false,
  });

  const post = await actions.getActivePostWithRhsField({
    id: firstPost.id,
  });

  expect(post!.id).toEqual(firstPost.id);
});

test("get action where expressions with M:1 relations with RHS field operand - publisher not active - model not returned", async () => {
  const publisher = await models.publisher.create({
    orgName: "Keel Org",
    isActive: false,
    booleanValue: true,
  });
  const author = await models.author.create({
    name: "Keelson",
    thePublisherId: publisher.id,
    isActive: true,
  });
  const firstPost = await models.post.create({
    title: "My First Post",
    theAuthorId: author.id,
    isActive: true,
  });

  const p = await actions.getActivePostWithRhsField({ id: firstPost.id });
  expect(p).toEqual(null);
});

test("get action where expressions with M:1 relations with RHS field operand - author not active - model not returned", async () => {
  const publisher = await models.publisher.create({
    orgName: "Keel Org",
    isActive: true,
    booleanValue: true,
  });
  const author = await models.author.create({
    name: "Keelson",
    thePublisherId: publisher.id,
    isActive: false,
  });
  const firstPost = await models.post.create({
    title: "My First Post",
    theAuthorId: author.id,
    isActive: true,
  });

  const p = await actions.getActivePostWithRhsField({ id: firstPost.id });
  expect(p).toEqual(null);
});

test("get action where expressions with M:1 relations with RHS field operand - post not active - model not returned", async () => {
  const publisher = await models.publisher.create({
    orgName: "Keel Org",
    isActive: true,
    booleanValue: true,
  });
  const author = await models.author.create({
    name: "Keelson",
    thePublisherId: publisher.id,
    isActive: true,
  });
  const firstPost = await models.post.create({
    title: "My First Post",
    theAuthorId: author.id,
    isActive: false,
  });

  const p = await actions.getActivePostWithRhsField({ id: firstPost.id });
  expect(p).toEqual(null);
});

test("get action where expressions with 1:M relations with RHS field operand - all models active - publisher returned", async () => {
  const publisherKeel = await models.publisher.create({
    orgName: "Keel Org",
    isActive: true,
    booleanValue: true,
  });
  const author1 = await models.author.create({
    name: "Keelson",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const publisher = await actions.getActivePublisherWithActivePostsWithRhsField(
    {
      id: publisherKeel.id,
    }
  );

  expect(publisher!.id).toEqual(publisherKeel.id);
});

test("get action where expressions with 1:M relations with RHS field operand - one active author - publisher returned", async () => {
  const publisherKeel = await models.publisher.create({
    orgName: "Keel Org",
    isActive: true,
    booleanValue: true,
  });
  const author1 = await models.author.create({
    name: "Keelson",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    thePublisherId: publisherKeel.id,
    isActive: false,
  });
  await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const publisher = await actions.getActivePublisherWithActivePostsWithRhsField(
    {
      id: publisherKeel.id,
    }
  );

  expect(publisher!.id).toEqual(publisherKeel.id);
});

test("get action where expressions with 1:M relations with RHS field operand - no active author - publisher not returned", async () => {
  const publisherKeel = await models.publisher.create({
    orgName: "Keel Org",
    isActive: true,
    booleanValue: true,
  });
  const author1 = await models.author.create({
    name: "Keelson",
    thePublisherId: publisherKeel.id,
    isActive: false,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    thePublisherId: publisherKeel.id,
    isActive: false,
  });
  await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const p = await actions.getActivePublisherWithActivePostsWithRhsField({
    id: publisherKeel.id,
  });
  expect(p).toEqual(null);
});

test("get action where expressions with 1:M relations with RHS field operand - one active post - publisher returned", async () => {
  const publisherKeel = await models.publisher.create({
    orgName: "Keel Org",
    isActive: true,
    booleanValue: true,
  });
  const author1 = await models.author.create({
    name: "Keelson",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: false,
  });
  await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: false,
  });

  const publisher = await actions.getActivePublisherWithActivePostsWithRhsField(
    {
      id: publisherKeel.id,
    }
  );

  expect(publisher!.id).toEqual(publisherKeel.id);
});

test("get action where expressions with 1:M relations with RHS field operand - no active posts - publisher not returned", async () => {
  const publisherKeel = await models.publisher.create({
    orgName: "Keel Org",
    isActive: true,
    booleanValue: true,
  });
  const author1 = await models.author.create({
    name: "Keelson",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: false,
  });
  await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: false,
  });
  await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: false,
  });

  const p = await actions.getActivePublisherWithActivePostsWithRhsField({
    id: publisherKeel.id,
  });
  expect(p).toEqual(null);
});

test("list action where expressions with M:1 relations with RHS field operand - all models active - all models returned", async () => {
  const publisherKeel = await models.publisher.create({
    orgName: "Keel Org",
    isActive: true,
    booleanValue: true,
  });
  const publisherWeave = await models.publisher.create({
    orgName: "Weave Org",
    isActive: true,
    booleanValue: true,
  });
  const author1 = await models.author.create({
    name: "Keelson",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    thePublisherId: publisherWeave.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const { results: posts } = await actions.listActivePostsWithRhsField({});

  expect(posts.length).toEqual(3);
});

test("list action where expressions with M:1 relations with RHS field operand - matching active status - all models returned", async () => {
  const publisherKeel = await models.publisher.create({
    orgName: "Keel Org",
    isActive: true,
    booleanValue: true,
  });
  const publisherWeave = await models.publisher.create({
    orgName: "Weave Org",
    isActive: false,
    booleanValue: false,
  });
  const author1 = await models.author.create({
    name: "Keelson",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    thePublisherId: publisherWeave.id,
    isActive: false,
  });
  await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: false,
  });

  const { results: posts } = await actions.listActivePostsWithRhsField({});

  expect(posts.length).toEqual(3);
});

test("list action where expressions with M:1 relations with RHS field operand - one active author - Keelson posts returned", async () => {
  const publisherKeel = await models.publisher.create({
    orgName: "Keel Org",
    isActive: true,
    booleanValue: true,
  });
  const publisherWeave = await models.publisher.create({
    orgName: "Weave Org",
    isActive: true,
    booleanValue: true,
  });
  const author1 = await models.author.create({
    name: "Keelson",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    thePublisherId: publisherWeave.id,
    isActive: false,
  });
  await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const { results: posts } = await actions.listActivePostsWithRhsField({});

  expect(posts.length).toEqual(2);
});

test("list action where expressions with M:1 relations with RHS field operand - Weaveton author inactive and one active Keelson post - other Keelson post returned", async () => {
  const publisherKeel = await models.publisher.create({
    orgName: "Keel Org",
    isActive: true,
    booleanValue: true,
  });
  const publisherWeave = await models.publisher.create({
    orgName: "Weave Org",
    isActive: true,
    booleanValue: true,
  });
  const author1 = await models.author.create({
    name: "Keelson",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    thePublisherId: publisherWeave.id,
    isActive: false,
  });
  await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: false,
  });
  await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const { results: posts } = await actions.listActivePostsWithRhsField({});

  expect(posts.length).toEqual(1);
});

test("list action where expressions with 1:M relations with RHS field operand - all models active - all models returned", async () => {
  const publisherKeel = await models.publisher.create({
    orgName: "Keel Org 2",
    isActive: true,
    booleanValue: true,
  });
  const publisherWeave = await models.publisher.create({
    orgName: "Weave Org 2",
    isActive: true,
    booleanValue: true,
  });
  const author1 = await models.author.create({
    name: "Keelson",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    thePublisherId: publisherWeave.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const { results: publishers } =
    await actions.listActivePublishersWithActivePostsWithRhsField({});

  expect(publishers.length).toEqual(2);
});

test("list action where expressions with 1:M relations with RHS field operand - Weaveton post inactive - Keelson publisher returned", async () => {
  const publisherKeel = await models.publisher.create({
    orgName: "Keel Org 2",
    isActive: true,
    booleanValue: true,
  });
  const publisherWeave = await models.publisher.create({
    orgName: "Weave Org 2",
    isActive: true,
    booleanValue: true,
  });
  const author1 = await models.author.create({
    name: "Keelson",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    thePublisherId: publisherWeave.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: false,
  });

  const { results: publishers } =
    await actions.listActivePublishersWithActivePostsWithRhsField({});

  expect(publishers.length).toEqual(1);
});

test("list action where expressions with 1:M relations with RHS field operand - only one Keelson post inactive - all publishers returned", async () => {
  const publisherKeel = await models.publisher.create({
    orgName: "Keel Org 2",
    isActive: true,
    booleanValue: true,
  });
  const publisherWeave = await models.publisher.create({
    orgName: "Weave Org 2",
    isActive: true,
    booleanValue: true,
  });
  const author1 = await models.author.create({
    name: "Keelson",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    thePublisherId: publisherWeave.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: false,
  });
  await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const { results: publishers } =
    await actions.listActivePublishersWithActivePostsWithRhsField({});

  expect(publishers.length).toEqual(2);
});

test("list action where expressions with 1:M relations with RHS field operand - Keelson author inactive and Weaveton post inactive - no publishers returned", async () => {
  const publisherKeel = await models.publisher.create({
    orgName: "Keel Org 2",
    isActive: true,
    booleanValue: true,
  });
  const publisherWeave = await models.publisher.create({
    orgName: "Weave Org 2",
    isActive: true,
    booleanValue: true,
  });
  const author1 = await models.author.create({
    name: "Keelson",
    thePublisherId: publisherKeel.id,
    isActive: false,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    thePublisherId: publisherWeave.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: false,
  });

  const { results: publishers } =
    await actions.listActivePublishersWithActivePostsWithRhsField({});

  expect(publishers.length).toEqual(0);
});

test("where expressions which references models multiple times - Keel has active posts, Weave has no active posts - Keel post returned, Weave not returned", async () => {
  const publisherKeel = await models.publisher.create({
    orgName: "Keel Org",
    isActive: true,
  });
  const publisherWeave = await models.publisher.create({
    orgName: "Weave Org",
    isActive: true,
  });
  const author1 = await models.author.create({
    name: "Keelson",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    thePublisherId: publisherWeave.id,
    isActive: true,
  });
  const post1 = await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: false,
  });
  const post3 = await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: false,
  });
  await models.post.create({
    title: "Weaveton Second Post",
    theAuthorId: author2.id,
    isActive: false,
  });

  const post = await actions.getPostModelsReferencedMoreThanOnce({
    id: post1.id,
  });

  expect(post!.id).toEqual(post1.id);

  const p = await actions.getPostModelsReferencedMoreThanOnce({ id: post3.id });
  expect(p).toEqual(null);
});

test("delete action where expressions with M:1 relations - all models active - model deleted", async () => {
  const publisher = await models.publisher.create({
    orgName: "Keel Org",
    isActive: true,
  });
  const author = await models.author.create({
    name: "Keelson",
    thePublisherId: publisher.id,
    isActive: true,
  });
  const firstPost = await models.post.create({
    title: "My First Post",
    theAuthorId: author.id,
    isActive: true,
  });

  const deleted = await actions.deleteActivePost({ id: firstPost.id });

  expect(deleted).toEqual(firstPost.id);
});

test("delete action where expressions with M:1 relations - post model not active - no records found", async () => {
  const publisher = await models.publisher.create({
    orgName: "Keel Org",
    isActive: true,
  });
  const author = await models.author.create({
    name: "Keelson",
    thePublisherId: publisher.id,
    isActive: true,
  });
  const firstPost = await models.post.create({
    title: "My First Post",
    theAuthorId: author.id,
    isActive: false,
  });

  await expect(
    actions.deleteActivePost({ id: firstPost.id })
  ).rejects.toMatchObject({
    code: "ERR_RECORD_NOT_FOUND",
    message: "record not found",
  });
});

test("delete action where expressions with M:1 relations - publisher model not active - no records found", async () => {
  const publisher = await models.publisher.create({
    orgName: "Keel Org",
    isActive: false,
  });
  const author = await models.author.create({
    name: "Keelson",
    thePublisherId: publisher.id,
    isActive: true,
  });
  const firstPost = await models.post.create({
    title: "My First Post",
    theAuthorId: author.id,
    isActive: true,
  });

  await expect(
    actions.deleteActivePost({ id: firstPost.id })
  ).rejects.toMatchObject({
    code: "ERR_RECORD_NOT_FOUND",
    message: "record not found",
  });
});

test("delete action where expressions with 1:M relations - all models active - publisher deleted", async () => {
  const publisherKeel = await models.publisher.create({
    orgName: "Keel Org",
    isActive: true,
  });
  const author1 = await models.author.create({
    name: "Keelson",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const deleted = await actions.deleteActivePublisherWithActivePosts({
    id: publisherKeel.id,
  });

  expect(deleted).toEqual(publisherKeel.id);
});

test("delete action where expressions with 1:M relations - publisher not active - no publisher found", async () => {
  const publisherKeel = await models.publisher.create({
    orgName: "Keel Org",
    isActive: false,
  });
  const author1 = await models.author.create({
    name: "Keelson",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  await expect(
    actions.deleteActivePublisherWithActivePosts({ id: publisherKeel.id })
  ).rejects.toMatchObject({
    code: "ERR_RECORD_NOT_FOUND",
    message: "record not found",
  });
});

test("delete action where expressions with 1:M relations - single post active - publisher deleted", async () => {
  const publisherKeel = await models.publisher.create({
    orgName: "Keel Org",
    isActive: true,
  });
  const author1 = await models.author.create({
    name: "Keelson",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: false,
  });
  await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: false,
  });
  await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const deleted = await actions.deleteActivePublisherWithActivePosts({
    id: publisherKeel.id,
  });

  expect(deleted).toEqual(publisherKeel.id);
});

test("delete action where expressions with 1:M relations - posts not active - no publisher found", async () => {
  const publisherKeel = await models.publisher.create({
    orgName: "Keel Org",
    isActive: true,
  });
  const author1 = await models.author.create({
    name: "Keelson",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  const author2 = await models.author.create({
    name: "Weaveton",
    thePublisherId: publisherKeel.id,
    isActive: true,
  });
  await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: false,
  });
  await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: false,
  });
  await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: false,
  });

  await expect(
    actions.deleteActivePublisherWithActivePosts({ id: publisherKeel.id })
  ).rejects.toMatchObject({
    code: "ERR_RECORD_NOT_FOUND",
    message: "record not found",
  });
});

test("get action where expressions with M:1 relations - depends on @relation", async () => {
  const publisher = await models.publisher.create({
    orgName: "Keel Org",
    isActive: true,
  });
  const scribe = await models.author.create({
    name: "scribe42",
    thePublisherId: publisher.id,
    isActive: true,
  });
  const reviewerJane = await models.author.create({
    name: "reviewerJane",
    thePublisherId: publisher.id,
    isActive: true,
  });
  const reviewerJohn = await models.author.create({
    name: "reviewerJohn",
    thePublisherId: publisher.id,
    isActive: true,
  });
  const postReviewedByJane = await models.post.create({
    title: "unused",
    theAuthorId: scribe.id,
    theReviewerId: reviewerJane.id,
    isActive: true,
  });
  const postReviewedByJohn = await models.post.create({
    title: "unused",
    theAuthorId: scribe.id,
    theReviewerId: reviewerJohn.id,
    isActive: true,
  });

  const { results: reviewers } = await actions.listReviewerByPostId({
    where: {
      reviewedPosts: { id: { equals: postReviewedByJane.id } },
    },
  });

  expect(reviewers[0]!.name).toEqual("reviewerJane");
});
