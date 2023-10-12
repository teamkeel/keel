import { actions, models, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

test("relationships and backlinks - get", async () => {
  const maryIdentity = await models.identity.create({
    email: "mary@weave.so ",
  });
  const mary = await actions
    .withIdentity(maryIdentity)
    .newAccount({ username: "mary77" });

  const result = await actions.withIdentity(maryIdentity).getMyAccount();
  expect(result).not.toBeNull();
  expect(result!.id).toEqual(mary.id);
});

test("relationships and backlinks - in operator", async () => {
  const maryIdentity = await models.identity.create({
    email: "mary@weave.so ",
  });
  const johnIdentity = await models.identity.create({
    email: "john@weave.so ",
  });
  const arthurIdentity = await models.identity.create({
    email: "arthur@weave.so ",
  });

  const mary = await actions
    .withIdentity(maryIdentity)
    .newAccount({ username: "mary77" });
  const john = await actions
    .withIdentity(johnIdentity)
    .newAccount({ username: "johndoe" });
  const arthur = await actions
    .withIdentity(arthurIdentity)
    .newAccount({ username: "art" });

  // John follows Mary
  await actions.follow({
    followee: { id: mary.id },
    follower: { id: john.id },
  });

  // John follows Arthur
  await actions.follow({
    followee: { id: arthur.id },
    follower: { id: john.id },
  });

  // Arthur follows Mary
  await actions.follow({
    followee: { id: mary.id },
    follower: { id: arthur.id },
  });

  const johnIsFollowing = await actions
    .withIdentity(johnIdentity)
    .accountsFollowed();
  expect(johnIsFollowing.results).toHaveLength(2);
  expect(johnIsFollowing.results[0].username).toEqual("art");
  expect(johnIsFollowing.results[1].username).toEqual("mary77");

  const arthurIsFollowing = await actions
    .withIdentity(arthurIdentity)
    .accountsFollowed();
  expect(arthurIsFollowing.results).toHaveLength(1);
  expect(arthurIsFollowing.results[0].username).toEqual("mary77");

  const maryIsFollowing = await actions
    .withIdentity(maryIdentity)
    .accountsFollowed();
  expect(maryIsFollowing.results).toHaveLength(0);

  const followingJohn = await actions
    .withIdentity(johnIdentity)
    .accountsFollowingMe();
  expect(followingJohn.results).toHaveLength(0);

  const followingArthur = await actions
    .withIdentity(arthurIdentity)
    .accountsFollowingMe();
  expect(followingArthur.results).toHaveLength(1);
  expect(followingArthur.results[0].username).toEqual("johndoe");

  const followingMary = await actions
    .withIdentity(maryIdentity)
    .accountsFollowingMe();
  expect(followingMary.results).toHaveLength(2);
  expect(followingMary.results[0].username).toEqual("art");
  expect(followingMary.results[1].username).toEqual("johndoe");
});

test("relationships and backlinks - not in operator", async () => {
  const maryIdentity = await models.identity.create({
    email: "mary@weave.so ",
  });
  const johnIdentity = await models.identity.create({
    email: "john@weave.so ",
  });
  const arthurIdentity = await models.identity.create({
    email: "arthur@weave.so ",
  });

  const mary = await actions
    .withIdentity(maryIdentity)
    .newAccount({ username: "mary77" });
  const john = await actions
    .withIdentity(johnIdentity)
    .newAccount({ username: "johndoe" });
  const arthur = await actions
    .withIdentity(arthurIdentity)
    .newAccount({ username: "art" });

  // John follows Mary
  await actions.follow({
    followee: { id: mary.id },
    follower: { id: john.id },
  });

  // John follows Arthur
  await actions.follow({
    followee: { id: arthur.id },
    follower: { id: john.id },
  });

  // Arthur follows Mary
  await actions.follow({
    followee: { id: mary.id },
    follower: { id: arthur.id },
  });

  const johnIsNotFollowing = await actions
    .withIdentity(johnIdentity)
    .accountsNotFollowed();
  expect(johnIsNotFollowing.results).toHaveLength(0);

  const arthurIsNotFollowing = await actions
    .withIdentity(arthurIdentity)
    .accountsNotFollowed();
  expect(arthurIsNotFollowing.results).toHaveLength(1);
  expect(arthurIsNotFollowing.results[0].username).toEqual("johndoe");

  const maryIsNotFollowing = await actions
    .withIdentity(maryIdentity)
    .accountsNotFollowed();
  expect(maryIsNotFollowing.results).toHaveLength(2);

  const notFollowingJohn = await actions
    .withIdentity(johnIdentity)
    .accountsNotFollowingMe();
  expect(notFollowingJohn.results).toHaveLength(2);
  expect(notFollowingJohn.results[0].username).toEqual("art");
  expect(notFollowingJohn.results[1].username).toEqual("mary77");

  const notFollowingArthur = await actions
    .withIdentity(arthurIdentity)
    .accountsNotFollowingMe();
  expect(notFollowingArthur.results).toHaveLength(1);
  expect(notFollowingArthur.results[0].username).toEqual("mary77");

  const notFollowingMary = await actions
    .withIdentity(maryIdentity)
    .accountsNotFollowingMe();
  expect(notFollowingMary.results).toHaveLength(0);
});
