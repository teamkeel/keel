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

test("updating a post with a permission rule", async () => {
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
    title: "a post",
    businessId: business.id,
  });

  await expect(
    actions.withIdentity(identity).updatePost({
      where: { id: post.id },
      values: {
        title: "changed post title",
      },
    })
  ).not.toHaveAuthorizationError();

  await expect(
    actions.withIdentity(otherIdentity).updatePost({
      where: { id: post.id },
      values: {
        title: "changed by an unauthorized user",
      },
    })
  ).toHaveAuthorizationError();

  const postFromDb = await models.post.findOne({ id: post.id });

  expect(postFromDb?.title).toEqual("changed post title");
});

test("deleting a post with a permission rule", async () => {
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
    title: "a post",
    businessId: business.id,
  });

  await expect(
    actions.withIdentity(identity).deletePost({
      id: post.id,
    })
  ).not.toHaveAuthorizationError();

  const post2 = await models.post.create({
    title: "a second post to be deleted",
    businessId: business.id,
  });

  await expect(
    actions.withIdentity(otherIdentity).deletePost({
      id: post2.id,
    })
  ).toHaveAuthorizationError();
});

test("creating a post with a role based permission rule - email based - permitted", async () => {
  const { token } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "verified@keel.xyz",
      password: "1234",
    },
  });

  await models.identity.update(
    {
      email: "verified@keel.xyz",
      issuer: "keel",
    },
    {
      emailVerified: true,
    }
  );

  const identity = await models.identity.create({
    email: "businessowner@keel.xyz",
    password: "foo",
  });
  const business = await models.business.create({
    name: "Adam Inc",
    identityId: identity.id,
  });

  await expect(
    actions.withAuthToken(token).createPostWithRole({
      title: "a post created via a special role",
      business: { id: business.id },
    })
  ).not.toHaveAuthorizationError();
});

test("creating a post with a role based permission rule - email based - unpermitted", async () => {
  const { token } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "disallowed@keel.xyz",
      password: "1234",
    },
  });

  const identity = await models.identity.create({
    email: "businessowner@keel.xyz",
    password: "foo",
  });
  const business = await models.business.create({
    name: "Adam Inc",
    identityId: identity.id,
  });

  await expect(
    actions.withAuthToken(token).createPostWithRole({
      title: "a post created via a special role",
      business: { id: business.id },
    })
  ).toHaveAuthorizationError();
});

test("creating a post with a role based permission rule - domain based - permitted", async () => {
  const { token } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "adam@abc.com",
      password: "1234",
    },
  });

  await models.identity.update(
    {
      email: "adam@abc.com",
      issuer: "keel",
    },
    {
      emailVerified: true,
    }
  );

  const identity = await models.identity.create({
    email: "businessowner@keel.xyz",
    password: "foo",
  });
  const business = await models.business.create({
    name: "Adam Inc",
    identityId: identity.id,
  });

  await expect(
    actions.withAuthToken(token).createPostWithRole({
      title: "a post created via a special role",
      business: { id: business.id },
    })
  ).not.toHaveAuthorizationError();
});

test("creating a post with a role based permission rule - domain based - unpermitted", async () => {
  const { token } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "blah@bca.com",
      password: "1234",
    },
  });

  const identity = await models.identity.create({
    email: "businessowner@keel.xyz",
    password: "foo",
  });
  const business = await models.business.create({
    name: "Adam Inc",
    identityId: identity.id,
  });

  await expect(
    actions.withAuthToken(token).createPostWithRole({
      title: "a post created via a special role",
      business: { id: business.id },
    })
  ).toHaveAuthorizationError();
});
