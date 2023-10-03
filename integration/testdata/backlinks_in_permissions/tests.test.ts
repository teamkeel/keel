import { actions, models, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

test("admit action - with adult identity - should be authorized", async () => {

  const adultIdentity = await models.identity.create({ email: "adults@someplace.com" });
  const childIdentity = await models.identity.create({ email: "child@someplace.com" });

  const adultUser = await models.user.create({ 
    isAdult: true,
    identityId: adultIdentity.id
  });

  const childUser = await models.user.create({ 
    isAdult: false,
    identityId: childIdentity.id
  });

  const film = await models.adultFilm.create({
    title: "some film"
  });

  await expect(
    actions.admit({
      film: { id: film.id },
      identity: { id: adultIdentity.id }
    })
  ).not.toHaveError({});

  await expect(
    actions.admit({
      film: { id: film.id },
      identity: { id: childIdentity.id }
    })
  ).toHaveAuthorizationError();
});

test("getFilm action - with adult user - should be authorized", async () => {

  const adultIdentity = await models.identity.create({ email: "adults@someplace.com" });
  const childIdentity = await models.identity.create({ email: "child@someplace.com" });

  const adultUser = await actions.createUser({
      isAdult: true,
      identity: {
        id: adultIdentity.id
      }});

  const childUser = await actions.createUser({
    isAdult: false,
    identity: {
      id: childIdentity.id
    }});

  await actions.createFilm({
    title: "some film"
  })

  // getFilm has a permission expression (which uses an Identity back-link):
  // ctx.identity.user.isAdult == true
  // Which should grant access to films only to Users who are adult.

  await actions.withIdentity(adultIdentity).getFilm({
    title: "some film"
  })

});