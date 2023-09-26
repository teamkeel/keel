import { actions, models } from "@teamkeel/testing";
import { test } from "vitest";

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

  // This permission expression (which uses an Identity back-link):
  // ctx.identity.user.isAdult == true
  // Should grant access to films only to Users who are adult.

  await actions.withIdentity(adultIdentity).getFilm({
    title: "some film"
  })

});