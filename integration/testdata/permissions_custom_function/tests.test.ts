import { actions, models, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

test("fetching (get) a post with a permission rule", async () => {
  const identity = await models.identity.create({
    email: "adam@keel.xyz",
    password: "foo",
  });

  const otherIdentity = await models.identity.create({
    email: "jon@keel.xyz",
    password: "foo",
  });

  const business = await models.business.create({
    name: "Adam Inc",
    identityId: identity.id,
  });

  const post = await models.post.create({
    title: "A post about " + business.name,
    businessId: business.id,
  });

  await expect(
    actions.withIdentity(otherIdentity).getPost({
      id: post.id,
    })
  ).toHaveAuthorizationError();

  await expect(
    actions.withIdentity(identity).getPost({
      id: post.id,
    })
  ).not.toHaveAuthorizationError();
});

test("getting a post with a value expression permission rule", async () => {
  const identity = await models.identity.create({
    email: "adam@keel.xyz",
    password: "foo",
  });

  const business = await models.business.create({
    name: "Adam Inc",
    identityId: identity.id,
  });

  const secretPost = await models.post.create({
    title: "A post about " + business.name,
    businessId: business.id,
  });

  // this will fail because no identity is attached to the request
  await expect(
    actions.getSecretPost({
      id: secretPost.id,
    })
  ).toHaveAuthorizationError();

  // this will succeed because the permission rule specifies that isAuthenticated must be true
  await expect(
    actions.withIdentity(identity).getSecretPost({
      id: secretPost.id,
    })
  ).not.toHaveAuthorizationError();
});

test("listing many posts with a permission rule - failing", async () => {
  const identity = await models.identity.create({
    email: "adam@keel.xyz",
    password: "foo",
  });

  const otherIdentity = await models.identity.create({
    email: "jon@keel.xyz",
    password: "foo",
  });

  const business = await models.business.create({
    name: "Adam Inc",
    identityId: identity.id,
  });

  await models.post.create({
    title: "A post about " + business.name,
    businessId: business.id,
  });

  const otherBusiness = await models.business.create({
    name: "Jon Inc",
    identityId: otherIdentity.id,
  });

  await models.post.create({
    title: "A post about another business",
    businessId: otherBusiness.id,
  });

  // we always expect there to be an auth error here
  // as the list function returns all posts and one post will be owned by the other business
  await expect(
    actions.withIdentity(identity).listPosts({})
  ).toHaveAuthorizationError();
});

test("listing many posts with a permission rule - succeeding", async () => {
  const identity = await models.identity.create({
    email: "adam@keel.xyz",
    password: "foo",
  });

  const business = await models.business.create({
    name: "Adam Inc",
    identityId: identity.id,
  });

  await models.post.create({
    title: "A post about " + business.name,
    businessId: business.id,
  });

  // given there are no posts in the db that are owned by another business
  // we expect there to be no auth error returned as the list function returns all posts
  // and one post will be owned by the other business
  await expect(
    actions.withIdentity(identity).listPosts({})
  ).not.toHaveAuthorizationError();
});

test("creating a post with a permission rule", async () => {
  const identity = await models.identity.create({
    email: "adam@keel.xyz",
    password: "foo",
  });

  const otherIdentity = await models.identity.create({
    email: "jon@keel.xyz",
    password: "foo",
  });

  const business = await models.business.create({
    name: "Adam Inc",
    identityId: identity.id,
  });

  await expect(
    actions.withIdentity(otherIdentity).createPost({
      title: "a post about " + business.name,
      business: { id: business.id },
    })
  ).toHaveAuthorizationError();

  await expect(
    actions.withIdentity(identity).createPost({
      title: "a post about " + business.name,
      business: { id: business.id },
    })
  ).not.toHaveAuthorizationError();
});
