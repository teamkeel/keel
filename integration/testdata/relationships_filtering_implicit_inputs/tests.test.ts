import { actions, models, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

test("get action implicit inputs with M:1 relations - all models active - model returned", async () => {
  const publisher = await models.publisher.create({
    orgName: "Keel Org",
    isActive: true,
  });
  const author = await models.author.create({
    name: "Keelson",
    thePublisherId: publisher.id,
    isActive: true,
  });
  const firstpost = await models.post.create({
    title: "My First Post",
    theAuthorId: author.id,
    isActive: true,
  });

  const post = await actions.getActivePost({
    id: firstpost.id,
    theAuthorThePublisherIsActive: true,
    theAuthorIsActive: true,
    isActive: true,
  });

  expect(post!.id).toEqual(firstpost.id);
});

test("get action implicit inputs with M:1 relations - post model not active - no records found", async () => {
  const publisher = await models.publisher.create({
    orgName: "Keel Org",
    isActive: true,
  });
  const author = await models.author.create({
    name: "Keelson",
    thePublisherId: publisher.id,
    isActive: true,
  });
  const firstpost = await models.post.create({
    title: "My First Post",
    theAuthorId: author.id,
    isActive: false,
  });

  expect(
    await actions.getActivePost({
      id: firstpost.id,
      theAuthorThePublisherIsActive: true,
      theAuthorIsActive: true,
      isActive: true,
    })
  ).toEqual(null);
});

test("get action implicit inputs with M:1 relations - nested author model not active - no records found", async () => {
  const publisher = await models.publisher.create({
    orgName: "Keel Org",
    isActive: true,
  });
  const author = await models.author.create({
    name: "Keelson",
    thePublisherId: publisher.id,
    isActive: false,
  });
  const firstpost = await models.post.create({
    title: "My First Post",
    theAuthorId: author.id,
    isActive: true,
  });

  expect(
    await actions.getActivePost({
      id: firstpost.id,
      theAuthorThePublisherIsActive: true,
      theAuthorIsActive: true,
      isActive: true,
    })
  ).toEqual(null);
});

test("get action implicit inputs with M:1 relations - nested nested publisher model not active - no records found", async () => {
  const publisher = await models.publisher.create({
    orgName: "Keel Org",
    isActive: false,
  });
  const author = await models.author.create({
    name: "Keelson",
    thePublisherId: publisher.id,
    isActive: true,
  });
  const firstpost = await models.post.create({
    title: "My First Post",
    theAuthorId: author.id,
    isActive: true,
  });

  expect(
    await actions.getActivePost({
      id: firstpost.id,
      theAuthorThePublisherIsActive: true,
      theAuthorIsActive: true,
      isActive: true,
    })
  ).toEqual(null);
});

test("get action implicit inputs with 1:M relations - all models active - publisher returned", async () => {
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
  const post1 = await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  const post2 = await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  const post3 = await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const publisher = await actions.getActivePublisherWithActivePosts({
    id: publisherKeel.id,
    theAuthorsThePostsIsActive: true,
    theAuthorsIsActive: true,
    isActive: true,
  });

  expect(publisher!.id).toEqual(publisherKeel.id);
});

test("get action implicit inputs with 1:M relations - publisher not active - no publisher found", async () => {
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
  const post1 = await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  const post2 = await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  const post3 = await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  expect(
    await actions.getActivePublisherWithActivePosts({
      id: publisherKeel.id,
      theAuthorsThePostsIsActive: true,
      theAuthorsIsActive: true,
      isActive: true,
    })
  ).toEqual(null);
});

test("get action implicit inputs with 1:M relations - one author active - publisher returned", async () => {
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
  const post1 = await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  const post2 = await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  const post3 = await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const publisher = await actions.getActivePublisherWithActivePosts({
    id: publisherKeel.id,
    theAuthorsThePostsIsActive: true,
    theAuthorsIsActive: true,
    isActive: true,
  });

  expect(publisher!.id).toEqual(publisherKeel.id);
});

test("get action implicit inputs with 1:M relations - active author with inactive posts and inactive autor with active posts - no publisher found", async () => {
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
  const post1 = await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: false,
  });
  const post2 = await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: false,
  });
  const post3 = await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  expect(
    await actions.getActivePublisherWithActivePosts({
      id: publisherKeel.id,
      theAuthorsThePostsIsActive: true,
      theAuthorsIsActive: true,
      isActive: true,
    })
  ).toEqual(null);
});

test("get action implicit inputs with 1:M relations - no active posts  - publisher returned", async () => {
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
  const post1 = await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: false,
  });
  const post2 = await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: false,
  });
  const post3 = await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: false,
  });

  expect(
    await actions.getActivePublisherWithActivePosts({
      id: publisherKeel.id,
      theAuthorsThePostsIsActive: true,
      theAuthorsIsActive: true,
      isActive: true,
    })
  ).toEqual(null);
});

test("list action implicit inputs with M:1 relations - all models active - all models returned", async () => {
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
  const post2 = await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  const post3 = await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const { results: posts } = await actions.listActivePosts({
    where: {
      theAuthor: {
        thePublisher: { isActive: { equals: true } },
        isActive: { equals: true },
      },
      isActive: { equals: true },
    },
  });

  expect(posts.length).toEqual(3);
});

test("list action implicit inputs with M:1 relations - Keel org not active - Weave models returned", async () => {
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
  const post1 = await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  const post2 = await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  const post3 = await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const { results: posts } = await actions.listActivePosts({
    where: {
      theAuthor: {
        thePublisher: { isActive: { equals: true } },
        isActive: { equals: true },
      },
      isActive: { equals: true },
    },
  });

  expect(posts.length).toEqual(1);
});

test("list action implicit inputs with M:1 relations - Keelson author not active - Weaveton models returned", async () => {
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
  const post1 = await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  const post2 = await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  const post3 = await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const { results: posts } = await actions.listActivePosts({
    where: {
      theAuthor: {
        thePublisher: { isActive: { equals: true } },
        isActive: { equals: true },
      },
      isActive: { equals: true },
    },
  });

  expect(posts.length).toEqual(1);
});

test("list action implicit inputs with M:1 relations - one Keelson post not active - Weaveton models returned", async () => {
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
    isActive: false,
  });
  const post2 = await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  const post3 = await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const { results: posts } = await actions.listActivePosts({
    where: {
      theAuthor: {
        thePublisher: { isActive: { equals: true } },
        isActive: { equals: true },
      },
      isActive: { equals: true },
    },
  });

  expect(posts.length).toEqual(2);
});

test("list action implicit inputs with 1:M relations - all models active - all models returned", async () => {
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
  const post1 = await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  const post2 = await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  const post3 = await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const { results: publishers } =
    await actions.listActivePublishersWithActivePosts({
      where: {
        theAuthors: {
          thePosts: { isActive: { equals: true } },
          isActive: { equals: true },
        },
        isActive: { equals: true },
      },
    });

  expect(publishers.length).toEqual(2);
});

test("list action implicit inputs with 1:M relations - Keel org not active - only Keel returned", async () => {
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
  const post1 = await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  const post2 = await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  const post3 = await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const { results: publishers } =
    await actions.listActivePublishersWithActivePosts({
      where: {
        theAuthors: {
          thePosts: { isActive: { equals: true } },
          isActive: { equals: true },
        },
        isActive: { equals: true },
      },
    });

  expect(publishers.length).toEqual(1);
});

test("list action implicit inputs with 1:M relations - Keel author not active - Weave org returned", async () => {
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
  const post1 = await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  const post2 = await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  const post3 = await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const { results: publishers } =
    await actions.listActivePublishersWithActivePosts({
      where: {
        theAuthors: {
          thePosts: { isActive: { equals: true } },
          isActive: { equals: true },
        },
        isActive: { equals: true },
      },
    });

  expect(publishers.length).toEqual(1);
});

test("list action implicit inputs with 1:M relations - one Keel post not active - all models returned", async () => {
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
  const post2 = await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: false,
  });
  const post3 = await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const { results: publishers } =
    await actions.listActivePublishersWithActivePosts({
      where: {
        theAuthors: {
          thePosts: { isActive: { equals: true } },
          isActive: { equals: true },
        },
        isActive: { equals: true },
      },
    });

  expect(publishers.length).toEqual(2);
});

test("list action implicit inputs with 1:M relations - all Keel posts not active - Weave org returned", async () => {
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
    isActive: false,
  });
  const post2 = await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: false,
  });
  const post3 = await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const { results: publishers } =
    await actions.listActivePublishersWithActivePosts({
      where: {
        theAuthors: {
          thePosts: { isActive: { equals: true } },
          isActive: { equals: true },
        },
        isActive: { equals: true },
      },
    });

  expect(publishers.length).toEqual(1);
});

test("implicit inputs which references models multiple times - Keel has active posts, Weave has no active posts - Keel post returned, Weave not returned", async () => {
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
  const post2 = await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: false,
  });
  const post3 = await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: false,
  });
  const post4 = await models.post.create({
    title: "Weaveton Second Post",
    theAuthorId: author2.id,
    isActive: false,
  });

  const post = await actions.getPostModelsReferencedMoreThanOnce({
    id: post1.id,
    theAuthorThePublisherTheAuthorsThePostsIsActive: true,
  });

  expect(post!.id).toEqual(post1.id);

  expect(
    await actions.getPostModelsReferencedMoreThanOnce({
      id: post3.id,
      theAuthorThePublisherTheAuthorsThePostsIsActive: true,
    })
  ).toEqual(null);
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
  const firstpost = await models.post.create({
    title: "My First Post",
    theAuthorId: author.id,
    isActive: true,
  });

  const deletedId = await actions.deleteActivePost({
    id: firstpost.id,
    theAuthorThePublisherIsActive: true,
    theAuthorIsActive: true,
    isActive: true,
  });

  expect(deletedId).toEqual(firstpost.id);
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
  const firstpost = await models.post.create({
    title: "My First Post",
    theAuthorId: author.id,
    isActive: false,
  });

  await expect(
    actions.deleteActivePost({
      id: firstpost.id,
      theAuthorThePublisherIsActive: true,
      theAuthorIsActive: true,
      isActive: true,
    })
  ).toHaveError({
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
  const firstpost = await models.post.create({
    title: "My First Post",
    theAuthorId: author.id,
    isActive: true,
  });

  await expect(
    actions.deleteActivePost({
      id: firstpost.id,
      theAuthorThePublisherIsActive: true,
      theAuthorIsActive: true,
      isActive: true,
    })
  ).toHaveError({
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

  const deletedId = await actions.deleteActivePublisherWithActivePosts({
    id: publisherKeel.id,
    theAuthorsThePostsIsActive: true,
    theAuthorsIsActive: true,
    isActive: true,
  });

  expect(deletedId).toEqual(publisherKeel.id);
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
  const post1 = await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  const post2 = await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: true,
  });
  const post3 = await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  await expect(
    actions.deleteActivePublisherWithActivePosts({
      id: publisherKeel.id,
      theAuthorsThePostsIsActive: true,
      theAuthorsIsActive: true,
      isActive: true,
    })
  ).toHaveError({
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
  const post1 = await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: false,
  });

  const post2 = await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: false,
  });

  const post3 = await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: true,
  });

  const deletedId = await actions.deleteActivePublisherWithActivePosts({
    id: publisherKeel.id,
    theAuthorsThePostsIsActive: true,
    theAuthorsIsActive: true,
    isActive: true,
  });

  expect(deletedId).toEqual(publisherKeel.id);
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
  const post1 = await models.post.create({
    title: "Keelson First Post",
    theAuthorId: author1.id,
    isActive: false,
  });
  const post2 = await models.post.create({
    title: "Keelson Second Post",
    theAuthorId: author1.id,
    isActive: false,
  });
  const post3 = await models.post.create({
    title: "Weaveton First Post",
    theAuthorId: author2.id,
    isActive: false,
  });

  await expect(
    actions.deleteActivePublisherWithActivePosts({
      id: publisherKeel.id,
      theAuthorsThePostsIsActive: true,
      theAuthorsIsActive: true,
      isActive: true,
    })
  ).toHaveError({
    message: "record not found",
  });
});
