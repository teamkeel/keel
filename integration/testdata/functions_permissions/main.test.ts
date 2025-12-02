import { Admission, Film, Identity, Post, Status } from "@teamkeel/sdk";
import { models, actions, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

// Helper to verify no hooks were executed
async function expectNoHookExecutions() {
  const logs = await models.hookLog.findMany();
  expect(logs.length).toEqual(0);
}

// Helper to verify specific hooks were executed
async function expectHookExecutions(expectedHooks: { actionName: string; hookName: string }[]) {
  const logs = await models.hookLog.findMany();
  expect(logs.length).toEqual(expectedHooks.length);
  for (const expected of expectedHooks) {
    const found = logs.find(
      (log) => log.actionName === expected.actionName && log.hookName === expected.hookName
    );
    expect(found).toBeTruthy();
  }
}

test("many-to-many - can only view users that are in a shared org", async () => {
  const meta = await models.organisation.create({
    name: "Meta",
  });
  const netflix = await models.organisation.create({
    name: "Netflix",
  });
  const microsoft = await models.organisation.create({
    name: "Microsoft",
  });

  const identityAdam = await models.identity.create({
    email: "adam@keel.xyz",
    password: "foo",
  });
  const adam = await models.user.create({
    name: "Adam",
    identityId: identityAdam.id,
  });

  const identityDave = await models.identity.create({
    email: "dave@keel.xyz",
    password: "foo",
  });
  const dave = await models.user.create({
    name: "Dave",
    identityId: identityDave.id,
  });

  const identityTom = await models.identity.create({
    email: "tom@keel.xyz",
    password: "foo",
  });
  const tom = await models.user.create({
    name: "Tom",
    identityId: identityTom.id,
  });

  // Adam work at Meta and Microsoft
  await models.userOrganisation.create({
    orgId: meta.id,
    userId: adam.id,
  });
  await models.userOrganisation.create({
    orgId: microsoft.id,
    userId: adam.id,
  });

  // Tom works at Meta
  await models.userOrganisation.create({
    orgId: meta.id,
    userId: tom.id,
  });

  // Dave works at Netflix and Microsoft
  await models.userOrganisation.create({
    orgId: netflix.id,
    userId: dave.id,
  });
  await models.userOrganisation.create({
    orgId: microsoft.id,
    userId: dave.id,
  });

  // Dave is trying to view the users of Microsoft, which are Dave + Adam.
  // This is allowed as Dave shared an org with both of these users
  const res = await actions.withIdentity(identityDave).listUsersByOrganisation({
    where: {
      orgs: {
        org: { id: { equals: microsoft.id } },
      },
    },
  });
  expect(res.results.length).toBe(2);

  // Dave is trying to view the users of Meta, which are Adam + Tom.
  // Dave is allowed to view Adam, as they both work at Microsoft
  // But Dave is NOT allowed to view Tom, as they do not work at any org together
  // Because there are records that fail the permission rule (Tom) permission will be denied
  await expect(
    actions.withIdentity(identityDave).listUsersByOrganisation({
      where: {
        orgs: {
          org: { id: { equals: meta.id } },
        },
      },
    })
  ).toHaveAuthorizationError();
});

test("boolean condition / multiple joins / >= condition", async () => {
  const pulpFiction = await models.film.create({
    title: "Pulp Fiction",
    ageRestriction: 18,
  });
  const barbie = await models.film.create({
    title: "Barbie",
    ageRestriction: 12,
  });
  const shrek = await models.film.create({
    title: "Shrek",
    ageRestriction: 0,
  });
  const dailyMail = await models.publication.create({
    name: "Daily Mail",
  });
  const timeout = await models.publication.create({
    name: "Timeout",
  });

  const bob = await models.identity.create({
    email: "bob@gmail.com",
  });
  await models.audience.create({
    isCritic: false,
    age: 22,
    identityId: bob.id,
  });

  const mike = await models.identity.create({
    email: "mike@gmail.com",
  });
  await models.audience.create({
    isCritic: false,
    age: 9,
    identityId: mike.id,
  });

  const sally = await models.identity.create({
    email: "sally@timeout.com",
  });
  await models.audience.create({
    isCritic: true,
    publicationId: timeout.id,
    age: 17,
    identityId: sally.id,
  });

  const kim = await models.identity.create({
    email: "kim@dailymail.com",
  });
  await models.audience.create({
    isCritic: true,
    publicationId: dailyMail.id,
    age: 15,
    identityId: kim.id,
  });

  const createAdmission = (i: Identity, f: Film) =>
    actions.withIdentity(i).createAdmission({
      film: {
        id: f.id,
      },
    });

  // Bob can watch Pulp Fiction because he is old enough
  await expect(createAdmission(bob, pulpFiction)).resolves.toBeTruthy();

  // Sally can watch Pulp Fiction because although she is not old enough she is a critic
  await expect(createAdmission(sally, pulpFiction)).resolves.toBeTruthy();

  // Kim cannot watch Pulp Fiction because although she is a critic she works for the Daily Mail
  await expect(createAdmission(kim, pulpFiction)).rejects.toEqual({
    code: "ERR_PERMISSION_DENIED",
    message: "not authorized to access",
  });

  // Kim can watch Barbie because she is old enough
  await expect(createAdmission(kim, barbie)).resolves.toBeTruthy();

  // Mike can watch Shrek because he is old enough
  await expect(createAdmission(mike, shrek)).resolves.toBeTruthy();

  // Mike cannot watch Barbie because he is too young
  await expect(createAdmission(mike, barbie)).rejects.toEqual({
    code: "ERR_PERMISSION_DENIED",
    message: "not authorized to access",
  });
});

// ==========================================
// SCHEMA PERMISSION FAILURE - NO HOOK EXECUTION TESTS
// ==========================================

test("create @function - schema permission failure prevents hook execution", async () => {
  // Create a user that we can reference as author
  const identity = await models.identity.create({
    email: "test@keel.xyz",
    password: "foo",
  });
  const user = await models.user.create({
    name: "Test User",
    identityId: identity.id,
  });

  // Call without authentication - should fail schema permission check
  await expect(
    actions.createPostWithHook({
      title: "Test Post",
      author: { id: user.id },
    })
  ).rejects.toEqual({
    code: "ERR_PERMISSION_DENIED",
    message: "not authorized to access",
  });

  // Verify no hooks were executed
  await expectNoHookExecutions();

  // Verify no post was created
  const posts = await models.post.findMany();
  expect(posts.length).toEqual(0);
});

test("create @function - schema permission success executes hooks", async () => {
  const identity = await models.identity.create({
    email: "test@keel.xyz",
    password: "foo",
  });
  const user = await models.user.create({
    name: "Test User",
    identityId: identity.id,
  });

  // Call with authentication - should pass schema permission check
  const post = await actions.withIdentity(identity).createPostWithHook({
    title: "Test Post",
    author: { id: user.id },
  });

  expect(post.title).toEqual("Test Post");

  // Verify hooks were executed
  await expectHookExecutions([
    { actionName: "createPostWithHook", hookName: "beforeWrite" },
    { actionName: "createPostWithHook", hookName: "afterWrite" },
  ]);
});

test("get @function - schema permission failure prevents hook execution", async () => {
  // Create a post directly in the database
  const identity = await models.identity.create({
    email: "test@keel.xyz",
    password: "foo",
  });
  const user = await models.user.create({
    name: "Test User",
    identityId: identity.id,
  });
  const post = await models.post.create({
    title: "Existing Post",
    authorId: user.id,
  });

  // Call without authentication - should fail schema permission check
  await expect(
    actions.getPostWithHook({ id: post.id })
  ).rejects.toEqual({
    code: "ERR_PERMISSION_DENIED",
    message: "not authorized to access",
  });

  // Verify no hooks were executed
  await expectNoHookExecutions();
});

test("get @function - schema permission success executes hooks", async () => {
  const identity = await models.identity.create({
    email: "test@keel.xyz",
    password: "foo",
  });
  const user = await models.user.create({
    name: "Test User",
    identityId: identity.id,
  });
  const post = await models.post.create({
    title: "Existing Post",
    authorId: user.id,
  });

  // Call with authentication - should pass schema permission check
  const result = await actions.withIdentity(identity).getPostWithHook({ id: post.id });

  expect(result!.title).toEqual("Existing Post");

  // Verify hooks were executed
  await expectHookExecutions([
    { actionName: "getPostWithHook", hookName: "beforeQuery" },
    { actionName: "getPostWithHook", hookName: "afterQuery" },
  ]);
});

test("list @function - schema permission failure prevents hook execution", async () => {
  // Create a post directly in the database
  const identity = await models.identity.create({
    email: "test@keel.xyz",
    password: "foo",
  });
  const user = await models.user.create({
    name: "Test User",
    identityId: identity.id,
  });
  await models.post.create({
    title: "Post 1",
    authorId: user.id,
  });
  await models.post.create({
    title: "Post 2",
    authorId: user.id,
  });

  // Call without authentication - should fail schema permission check
  await expect(actions.listPostsWithHook()).rejects.toEqual({
    code: "ERR_PERMISSION_DENIED",
    message: "not authorized to access",
  });

  // Verify no hooks were executed
  await expectNoHookExecutions();
});

test("list @function - schema permission success executes hooks", async () => {
  const identity = await models.identity.create({
    email: "test@keel.xyz",
    password: "foo",
  });
  const user = await models.user.create({
    name: "Test User",
    identityId: identity.id,
  });
  await models.post.create({
    title: "Post 1",
    authorId: user.id,
  });
  await models.post.create({
    title: "Post 2",
    authorId: user.id,
  });

  // Call with authentication - should pass schema permission check
  const result = await actions.withIdentity(identity).listPostsWithHook();

  expect(result.results.length).toEqual(2);

  // Verify hooks were executed
  await expectHookExecutions([
    { actionName: "listPostsWithHook", hookName: "beforeQuery" },
    { actionName: "listPostsWithHook", hookName: "afterQuery" },
  ]);
});

test("update @function - schema permission failure prevents hook execution", async () => {
  // Create a post directly in the database
  const identity = await models.identity.create({
    email: "test@keel.xyz",
    password: "foo",
  });
  const user = await models.user.create({
    name: "Test User",
    identityId: identity.id,
  });
  const post = await models.post.create({
    title: "Original Title",
    authorId: user.id,
  });

  // Call without authentication - should fail schema permission check
  await expect(
    actions.updatePostWithHook({
      where: { id: post.id },
      values: { title: "New Title" },
    })
  ).rejects.toEqual({
    code: "ERR_PERMISSION_DENIED",
    message: "not authorized to access",
  });

  // Verify no hooks were executed
  await expectNoHookExecutions();

  // Verify post was not updated
  const dbPost = await models.post.findOne({ id: post.id });
  expect(dbPost!.title).toEqual("Original Title");
});

test("update @function - schema permission success executes hooks", async () => {
  const identity = await models.identity.create({
    email: "test@keel.xyz",
    password: "foo",
  });
  const user = await models.user.create({
    name: "Test User",
    identityId: identity.id,
  });
  const post = await models.post.create({
    title: "Original Title",
    authorId: user.id,
  });

  // Call with authentication - should pass schema permission check
  const result = await actions.withIdentity(identity).updatePostWithHook({
    where: { id: post.id },
    values: { title: "New Title" },
  });

  expect(result.title).toEqual("New Title");

  // Verify hooks were executed
  await expectHookExecutions([
    { actionName: "updatePostWithHook", hookName: "beforeQuery" },
    { actionName: "updatePostWithHook", hookName: "beforeWrite" },
    { actionName: "updatePostWithHook", hookName: "afterWrite" },
  ]);
});

test("delete @function - schema permission failure prevents hook execution", async () => {
  // Create a post directly in the database
  const identity = await models.identity.create({
    email: "test@keel.xyz",
    password: "foo",
  });
  const user = await models.user.create({
    name: "Test User",
    identityId: identity.id,
  });
  const post = await models.post.create({
    title: "Post to Delete",
    authorId: user.id,
  });

  // Call without authentication - should fail schema permission check
  await expect(
    actions.deletePostWithHook({ id: post.id })
  ).rejects.toEqual({
    code: "ERR_PERMISSION_DENIED",
    message: "not authorized to access",
  });

  // Verify no hooks were executed
  await expectNoHookExecutions();

  // Verify post was not deleted
  const dbPost = await models.post.findOne({ id: post.id });
  expect(dbPost).not.toBeNull();
});

test("delete @function - schema permission success executes hooks", async () => {
  const identity = await models.identity.create({
    email: "test@keel.xyz",
    password: "foo",
  });
  const user = await models.user.create({
    name: "Test User",
    identityId: identity.id,
  });
  const post = await models.post.create({
    title: "Post to Delete",
    authorId: user.id,
  });

  // Call with authentication - should pass schema permission check
  const deletedId = await actions.withIdentity(identity).deletePostWithHook({ id: post.id });

  expect(deletedId).toEqual(post.id);

  // Verify hooks were executed
  await expectHookExecutions([
    { actionName: "deletePostWithHook", hookName: "beforeQuery" },
    { actionName: "deletePostWithHook", hookName: "beforeWrite" },
    { actionName: "deletePostWithHook", hookName: "afterWrite" },
  ]);

  // Verify post was deleted
  const dbPost = await models.post.findOne({ id: post.id });
  expect(dbPost).toBeNull();
});

test("enum literal in expression - can only delete users that are active", async () => {
  const identityActiveUser = await models.identity.create({
    email: "active@keel.xyz",
    password: "foo",
  });
  const activeUser = await models.user.create({
    name: "Active User",
    status: Status.Active,
    identityId: identityActiveUser.id,
  });

  const identityInactiveUser = await models.identity.create({
    email: "inactive@keel.xyz",
    password: "foo",
  });
  const inactiveUser = await models.user.create({
    name: "Inactive User",
    status: Status.Inactive,
    identityId: identityInactiveUser.id,
  });

  await expect(
    actions.withIdentity(identityActiveUser).deleteUser({ id: activeUser.id })
  ).resolves.toBeTruthy();
  await expect(
    actions
      .withIdentity(identityInactiveUser)
      .deleteUser({ id: inactiveUser.id })
  ).rejects.toEqual({
    code: "ERR_PERMISSION_DENIED",
    message: "not authorized to access",
  });
});
