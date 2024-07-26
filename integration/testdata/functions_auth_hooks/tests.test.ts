import { actions, resetDatabase, models } from "@teamkeel/testing";
import { beforeEach, expect, test } from "vitest";

beforeEach(resetDatabase);

test("test", async () => {
  const response1 = await fetch(
    process.env.KEEL_TESTING_AUTH_API_URL + "/token",
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        grant_type: "password",
        username: "user@keel.xyz",
        password: "1234",
      }),
    }
  );
  expect(response1.status).toEqual(200);

  const accounts1 = await models.account.findMany();
  expect(accounts1).toHaveLength(1);

  const account1 = accounts1[0];
  expect(account1.name).toEqual("Not set");
  expect(account1.loginCount).toEqual(1);

  const identity1 = await models.identity.findOne({ id: account1.identityId });
  expect(identity1?.email).toEqual("user@keel.xyz");

  const response2 = await fetch(
    process.env.KEEL_TESTING_AUTH_API_URL + "/token",
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        grant_type: "password",
        username: "user@keel.xyz",
        password: "wrong",
      }),
    }
  );
  expect(response2.status).toEqual(401);

  const accounts2 = await models.account.findMany();
  expect(accounts2).toHaveLength(1);

  const account2 = accounts2[0];
  expect(account2.name).toEqual("Not set");
  expect(account2.loginCount).toEqual(1);

  const identity2 = await models.identity.findOne({ id: account2.identityId });
  expect(identity2?.email).toEqual("user@keel.xyz");

  const response3 = await fetch(
    process.env.KEEL_TESTING_AUTH_API_URL + "/token",
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        grant_type: "password",
        username: "user@keel.xyz",
        password: "1234",
      }),
    }
  );
  expect(response3.status).toEqual(200);

  const accounts3 = await models.account.findMany();
  expect(accounts3).toHaveLength(1);

  const account3 = accounts3[0];
  expect(account3.name).toEqual("Not set");
  expect(account3.loginCount).toEqual(2);

  const identity3 = await models.identity.findOne({ id: account3.identityId });
  expect(identity3?.email).toEqual("user@keel.xyz");
});
