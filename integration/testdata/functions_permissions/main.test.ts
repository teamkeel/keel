import { Admission, Film, Identity } from "@teamkeel/sdk";
import { models, actions, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

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
    message: "not authorized to access this action",
  });

  // Kim can watch Barbie because she is old enough
  await expect(createAdmission(kim, barbie)).resolves.toBeTruthy();

  // Mike can watch Shrek because he is old enough
  await expect(createAdmission(mike, shrek)).resolves.toBeTruthy();

  // Mike cannot watch Barbie because he is too young
  await expect(createAdmission(mike, barbie)).rejects.toEqual({
    code: "ERR_PERMISSION_DENIED",
    message: "not authorized to access this action",
  });
});
