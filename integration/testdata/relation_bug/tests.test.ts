import { resetDatabase, models, actions } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

test("test", async () => {
  const identity = await models.identity.create({ email: "keelson@keel.so" });
  const user = await models.user.create({
    name: "Keelson",
    identityId: identity.id,
  });
  const team = await models.team.create({ name: "Team Keel" });

  const document = await models.document.create({
    title: "Road to Success",
    userId: user.id,
    teamId: team.id,
  });

  const documents = await actions.withIdentity(identity).listDocuments({
    where: {
      team: {
        id: {
          equals: team.id,
        },
      },
    },
  });

  console.log(documents.results);
  expect(documents.results).toHaveLength(1);
});
