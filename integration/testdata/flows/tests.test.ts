import { resetDatabase, models } from "@teamkeel/testing";
import { useDatabase } from "@teamkeel/sdk";
import { beforeEach, expect, test } from "vitest";
import { sql } from "kysely";

beforeEach(resetDatabase);

test("flows - basic execution", async () => {
  const token = await getToken();

  let { status, body } = await startFlow({
    name: "myFlow",
    token,
    body: {
      name: "Keelson",
      age: 23,
    },
  });
  expect(status).toEqual(200);
  const runId = body.id;

  const things = await models.thing.findMany();
  expect(things.length).toBe(1);
  expect(things[0].name).toBe("Keelson");

  const dbFlows = await sql`SELECT * FROM keel_flow_run`.execute(useDatabase());
  expect(dbFlows.rows.length).toBe(1);
  const dbSteps = await sql`SELECT * FROM keel_flow_step`.execute(
    useDatabase()
  );
  expect(dbSteps.rows.length).toBe(2);

  ({ status, body } = await getFlowRun({
    name: "myFlow",
    id: runId,
    token,
  }));

  expect(status).toEqual(200);
  expect(body.status).toBe("AWAITING_INPUT");

  const stepId = body.steps.find(
    (step) => step.type === "UI" && step.status === "PENDING"
  )?.id;

  ({ status, body } = await putStepValues({
    name: "myFlow",
    runId,
    stepId,
    token,
    values: {
      name: "Keelson updated",
      age: 32,
    },
  }));

  expect(status).toEqual(200);
  const newThings = await models.thing.findMany();
  expect(newThings.length).toBe(1);
  expect(newThings[0].name).toBe("Keelson updated");

  ({ status, body } = await getFlowRun({
    name: "myFlow",
    id: runId,
    token,
  }));
  expect(status).toEqual(200);
  expect(body.status).toBe("COMPLETED");
});

async function getToken() {
  const response = await fetch(
    process.env.KEEL_TESTING_AUTH_API_URL + "/token",
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        grant_type: "password",
        username: "admin@keel.xyz",
        password: "1234",
      }),
    }
  );
  expect(response.status).toEqual(200);

  const token = (await response.json()).access_token;
  await models.identity.update(
    {
      email: "admin@keel.xyz",
      issuer: "https://keel.so",
    },
    {
      emailVerified: true,
    }
  );

  return token;
}

async function startFlow({ name, token, body }) {
  const res = await fetch(
    `${process.env.KEEL_TESTING_API_URL}/flows/json/${name}`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: "Bearer " + token,
      },
      body: JSON.stringify(body),
    }
  );

  return {
    status: res.status,
    body: await res.json(),
  };
}

async function getFlowRun({ name, id, token }) {
  const res = await fetch(
    `${process.env.KEEL_TESTING_API_URL}/flows/json/${name}/${id}`,
    {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        Authorization: "Bearer " + token,
      },
    }
  );

  return {
    status: res.status,
    body: await res.json(),
  };
}

async function putStepValues({ name, runId, stepId, values, token }) {
  const res = await fetch(
    `${process.env.KEEL_TESTING_API_URL}/flows/json/${name}/${runId}/${stepId}`,
    {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
        Authorization: "Bearer " + token,
      },
      body: JSON.stringify(values),
    }
  );

  return {
    status: res.status,
    body: await res.json(),
  };
}
