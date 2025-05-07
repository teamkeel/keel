import { resetDatabase, models, flows } from "@teamkeel/testing";
import { useDatabase } from "@teamkeel/sdk";
import { beforeEach, expect, test } from "vitest";
import { sql } from "kysely";

beforeEach(resetDatabase);

test("flows - basic execution", async () => {
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

  const flowStartResponse = await fetch(
    `${process.env.KEEL_TESTING_API_URL}/flows/json/myFlow`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: "Bearer " + token,
      },
      body: JSON.stringify({
        name: "Keelson",
        age: 23,
      }),
    }
  );
  expect(flowStartResponse.status).toEqual(200);
  const flowStartData = await flowStartResponse.json();
  const runId = flowStartData.id;

  const things = await models.thing.findMany();
  expect(things.length).toBe(1);
  expect(things[0].name).toBe("Keelson");

  const dbFlows = await sql`SELECT * FROM keel_flow_run`.execute(useDatabase());
  expect(dbFlows.rows.length).toBe(1);
  const dbSteps = await sql`SELECT * FROM keel_flow_step`.execute(
    useDatabase()
  );
  expect(dbSteps.rows.length).toBe(2);

  const flowStateResponse = await fetch(
    `${process.env.KEEL_TESTING_API_URL}/flows/json/myFlow/${runId}`,
    {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        Authorization: "Bearer " + token,
      },
    }
  );
  expect(flowStateResponse.status).toEqual(200);
  const flowStateData = await flowStateResponse.json();
  expect(flowStateData.status).toBe("AWAITING_INPUT");

  const stepId = flowStateData.steps.find(
    (step) => step.type === "UI" && step.status === "PENDING"
  )?.id;

  const flowStepPutResponse = await fetch(
    `${process.env.KEEL_TESTING_API_URL}/flows/json/myFlow/${runId}/${stepId}`,
    {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
        Authorization: "Bearer " + token,
      },
      body: JSON.stringify({
        name: "Keelson updated",
        age: 32,
      }),
    }
  );
  expect(flowStepPutResponse.status).toEqual(200);
  const newThings = await models.thing.findMany();
  expect(newThings.length).toBe(1);
  expect(newThings[0].name).toBe("Keelson updated");

  const finalFlowStateResponse = await fetch(
    `${process.env.KEEL_TESTING_API_URL}/flows/json/myFlow/${runId}`,
    {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        Authorization: "Bearer " + token,
      },
    }
  );
  expect(finalFlowStateResponse.status).toEqual(200);
  const finalFlowStateData = await finalFlowStateResponse.json();
  expect(finalFlowStateData.status).toBe("COMPLETED");
});
