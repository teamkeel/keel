import { models, actions, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

test("two permissions rules, one includes fan out array", async () => {
  const identity = await models.identity.create({
    email: "adam@keel.xyz",
    password: "foo",
  });
  const user = await models.user.create({
    name: "Adam",
    identityId: identity.id,
  });

  await expect(
    actions.withIdentity(identity).getUser({
      id: user.id,
    })
  ).not.toHaveAuthorizationError();
});
