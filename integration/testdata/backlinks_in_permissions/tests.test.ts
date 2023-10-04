import { actions, models, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

test("admit with adult identity - permission backlink from ctx - is authorized", async () => {
  const adultIdentity = await models.identity.create({
    email: "adults@someplace.com",
  });

  await models.user.create({
    age: 18,
    identityId: adultIdentity.id,
  });

  const film = await models.film.create({
    title: "some film",
    ageRestriction: 16,
  });

  await expect(
    actions.admit({
      film: { id: film.id },
      identity: { id: adultIdentity.id },
    })
  ).not.toHaveError({});
});

test("admit with child identity - permission backlink from ctx - is not authorized", async () => {
  const childIdentity = await models.identity.create({
    email: "child@someplace.com",
  });

  await models.user.create({
    age: 12,
    identityId: childIdentity.id,
  });

  const film = await models.film.create({
    title: "some film",
    ageRestriction: 16,
  });

  await expect(
    actions.admit({
      film: { id: film.id },
      identity: { id: childIdentity.id },
    })
  ).toHaveAuthorizationError();
});

test("getFilm action with adult user - backlink from identity model field - is authorized", async () => {
  const adultIdentity = await models.identity.create({
    email: "adults@someplace.com",
  });

  await actions.createUser({
    age: 19,
    identity: {
      id: adultIdentity.id,
    },
  });

  await models.film.create({
    title: "some film",
    ageRestriction: 16,
  });

  await expect(
    actions.withIdentity(adultIdentity).getFilm({
      title: "some film",
    })
  ).not.toHaveError({});
});

test("getFilm action with child user - backlink from identity model field - is not authorized", async () => {
  const childIdentity = await models.identity.create({
    email: "child@someplace.com",
  });

  await actions.createUser({
    age: 12,
    identity: {
      id: childIdentity.id,
    },
  });

  await models.film.create({
    title: "some film",
    ageRestriction: 16,
  });

  await expect(
    actions.withIdentity(childIdentity).getFilm({
      title: "some film",
    })
  ).toHaveAuthorizationError();
});

// TODO: broken when there is no identity in ctx
// test("getFilm action with no user - backlink from identity model field - is not authorized", async () => {
//   await models.film.create({
//     title: "some film",
//     ageRestriction: 16,
//   });

//   await expect(
//     actions.getFilm({
//       title: "some film",
//     })
//   ).toHaveAuthorizationError();
// });

test("list members films - where attribute backlink from ctx - returns applicable films", async () => {
  await models.film.create({
    title: "Jurassic Park",
    ageRestriction: 16,
    onlyMembers: true,
  });

  await models.film.create({
    title: "Toy Story",
    ageRestriction: 12,
    onlyMembers: true,
  });

  await models.film.create({
    title: "Barney",
    ageRestriction: 2,
    onlyMembers: false,
  });

  await models.film.create({
    title: "The Matrix",
    ageRestriction: 16,
    onlyMembers: false,
  });

  const activeGroup = await models.membersGroup.create({
    name: "VIP",
  });
  const inActiveGroup = await models.membersGroup.create({
    name: "Inactive",
    isActive: false,
  });

  let identity = await models.identity.create({
    email: "young-member@someplace.com",
  });
  await models.user.create({
    age: 15,
    identityId: identity.id,
    groupId: activeGroup.id,
  });
  let films = await actions.withIdentity(identity).listMembersFilms();
  expect(films.results).length(2);
  expect(films.results[0].title).toEqual("Barney");
  expect(films.results[1].title).toEqual("Toy Story");

  identity = await models.identity.create({
    email: "young-deactived-member@someplace.com",
  });
  await models.user.create({
    age: 15,
    identityId: identity.id,
    groupId: inActiveGroup.id,
  });
  films = await actions.withIdentity(identity).listMembersFilms();
  expect(films.results).length(1);
  expect(films.results[0].title).toEqual("Barney");

  identity = await models.identity.create({
    email: "young-not-a-member@someplace.com",
  });
  await models.user.create({
    age: 15,
    identityId: identity.id,
  });
  films = await actions.withIdentity(identity).listMembersFilms();
  expect(films.results).length(1);
  expect(films.results[0].title).toEqual("Barney");

  identity = await models.identity.create({
    email: "adult-not-a-member@someplace.com",
  });
  await models.user.create({
    age: 18,
    identityId: identity.id,
  });
  films = await actions.withIdentity(identity).listMembersFilms();
  expect(films.results).length(2);
  expect(films.results[0].title).toEqual("Barney");
  expect(films.results[1].title).toEqual("The Matrix");
});
