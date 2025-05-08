import { resetDatabase, models } from "@teamkeel/testing";
import { beforeEach, expect, test } from "vitest";

beforeEach(resetDatabase);

/*
TEST CASES
========================================
[x] Stepless flow function
[ ] Flow function with consecutive UI steps
[ ] Flow function with consecutive function steps
[ ] Flow function with alternating function and UI steps
[ ] Error thrown in flow function
[ ] Error thrown in step function
[ ] Step returning scalar value
[ ] Step function retrying  
[ ] Step function timing out 
[ ] UI step validation
[ ] Check full API responses
[ ] All UI elements response
[ ] Test all Keel types as inputs
[ ] Permissions and identity tests


Should we test the full response from the API calls? (probably, since we're not testing this elsewhere)
Are we testing each UI element and all their properties?
Do we need to inspect the database in these tests? (probably not, since the API responses is mapped directly from the tables)
*/

test("flows - stepless flow", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  let { status, body } = await startFlow({
    name: "stepless",
    token,
    body: {},
  });
  expect(status).toEqual(200);

  const flow = await untilFlowFinished({
    name: "stepless",
    id: body.id,
    token,
  });
  expect(flow.status).toBe("COMPLETED");

  const things = await models.thing.findMany();
  expect(things.length).toBe(1);
  expect(things[0].name).toBe("Keelson");
});

test("flows - first step is a function", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  let { status, body } = await startFlow({
    name: "singleStep",
    token,
    body: {},
  });
  expect(status).toEqual(200);

  expect(body.status).toBe("RUNNING");
  expect(body.steps).toHaveLength(1);
  expect(body.steps[0].status).toBe("NEW");
  expect(body.steps[0].type).toBe("FUNCTION");
  expect(body.steps[0].output).toBeUndefined();

  const flow = await untilFlowFinished({
    name: "singleStep",
    id: body.id,
    token,
  });
  expect(flow.status).toBe("COMPLETED");
  expect(flow.steps[0].status).toBe("COMPLETED");
  expect(flow.steps).toHaveLength(1);
});

test("flows - alternating step types", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  let { status, body } = await startFlow({
    name: "MixedStepTypes",
    token,
    body: {
      name: "Keelson",
      age: 23,
    },
  });
  expect(status).toEqual(200);
  const runId = body.id;

  let flow = await untilFlowAwaitingInput({
    name: "MixedStepTypes",
    id: runId,
    token,
  });
  expect(flow.steps).toHaveLength(2);

  const things = await models.thing.findMany();
  expect(things.length).toBe(1);
  expect(things[0].name).toBe("Keelson");

  ({ status, body } = await getFlowRun({
    name: "MixedStepTypes",
    id: runId,
    token,
  }));

  expect(status).toEqual(200);
  expect(body.status).toBe("AWAITING_INPUT");

  const stepId = body.steps.find(
    (step) => step.type === "UI" && step.status === "PENDING"
  )?.id;

  ({ status, body } = await putStepValues({
    name: "MixedStepTypes",
    runId,
    stepId,
    token,
    values: {
      name: "Keelson updated",
      age: 32,
    },
  }));
  expect(status).toEqual(200);

  let stepData = body.steps.map((step) => ({
    name: step.name,
    status: step.status,
    type: step.type,
  }));
  expect(stepData).toEqual([
    { name: "insert thing", status: "COMPLETED", type: "FUNCTION" },
    { name: "confirm thing", status: "COMPLETED", type: "UI" },
    { name: "update thing", status: "NEW", type: "FUNCTION" },
  ]);

  flow = await untilFlowFinished({ name: "mixedStepTypes", id: runId, token });
  expect(flow.status).toBe("COMPLETED");

  stepData = flow.steps.map((step) => ({
    name: step.name,
    status: step.status,
    type: step.type,
  }));
  expect(stepData).toEqual([
    { name: "insert thing", status: "COMPLETED", type: "FUNCTION" },
    { name: "confirm thing", status: "COMPLETED", type: "UI" },
    { name: "update thing", status: "COMPLETED", type: "FUNCTION" },
  ]);

  const newThings = await models.thing.findMany();
  expect(newThings.length).toBe(1);
  expect(newThings[0].name).toBe("Keelson updated");
});

test("flows - authorised starting, getting and listing flows", async () => {
  const adminToken = await getToken({ email: "admin@keel.xyz" });
  const userToken = await getToken({ email: "user@gmail.com" });

  const resListAdmin = await listFlows({ token: adminToken });
  expect(resListAdmin.status).toBe(200);
  expect(resListAdmin.body.flows.length).toBe(3);
  expect(resListAdmin.body.flows[0].name).toBe("MixedStepTypes");
  expect(resListAdmin.body.flows[1].name).toBe("Stepless");

  const resListUser = await listFlows({ token: userToken });
  expect(resListUser.status).toBe(200);
  expect(resListUser.body.flows.length).toBe(1);
  expect(resListUser.body.flows[0].name).toBe("UserFlow");
});

test("flows - unauthorised starting flow", async () => {
  const token = await getToken({ email: "user@gmail.com" });
  const res = await startFlow({ name: "stepless", token, body: {} });
  expect(res.status).toBe(403);
});

test("flows - unauthenticated starting flow", async () => {
  const res = await startFlow({ name: "stepless", token: null, body: {} });
  expect(res.status).toBe(401);
});

test("flows - unauthorised getting flow", async () => {
  const adminToken = await getToken({ email: "admin@keel.xyz" });
  const resStart = await startFlow({
    name: "stepless",
    token: adminToken,
    body: {},
  });
  expect(resStart.status).toBe(200);

  const userToken = await getToken({ email: "user@gmail.com" });
  const resGet = await getFlowRun({
    name: "stepless",
    id: resStart.body.id,
    token: userToken,
  });
  expect(resGet.status).toBe(403);
});

test("flows - unauthenticated starting flow", async () => {
  const res = await startFlow({ name: "stepless", token: null, body: {} });
  expect(res.status).toBe(401);
});

test("flows - unauthenticated getting flow", async () => {
  const adminToken = await getToken({ email: "admin@keel.xyz" });
  const resStart = await startFlow({
    name: "stepless",
    token: adminToken,
    body: {},
  });
  expect(resStart.status).toBe(200);

  const resGet = await getFlowRun({
    name: "stepless",
    id: resStart.body.id,
    token: null,
  });
  expect(resGet.status).toBe(401);
});

test("flows - unauthenticated listing flows", async () => {
  const adminToken = await getToken({ email: "admin@keel.xyz" });
  const res1 = await startFlow({
    name: "stepless",
    token: adminToken,
    body: {},
  });
  const res2 = await startFlow({
    name: "stepless",
    token: adminToken,
    body: {},
  });
  expect(res1.status).toBe(200);
  expect(res2.status).toBe(200);

  const resGet = await listFlows({ token: null });
  expect(resGet.status).toBe(401);
});

async function getToken({ email }) {
  const response = await fetch(
    process.env.KEEL_TESTING_AUTH_API_URL + "/token",
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        grant_type: "password",
        username: email,
        password: "1234",
      }),
    }
  );
  expect(response.status).toEqual(200);

  const token = (await response.json()).access_token;
  await models.identity.update(
    {
      email: email,
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

async function listFlows({ token }) {
  const res = await fetch(`${process.env.KEEL_TESTING_API_URL}/flows/json`, {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
      Authorization: "Bearer " + token,
    },
  });

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

async function untilFlowAwaitingInput({ name, id, token }) {
  const startTime = Date.now();
  const timeout = 1000; // 1 seconds timeout on polling

  while (true) {
    if (Date.now() - startTime > timeout) {
      throw new Error(
        `timed out waiting for flow run to reach AWAITING_INPUT state after ${timeout}ms`
      );
    }

    const { status, body } = await getFlowRun({ name, id, token });
    expect(status).toEqual(200);

    if (body.status === "AWAITING_INPUT") {
      const lastStep = body.steps[body.steps.length - 1];
      expect(lastStep.status).toBe("PENDING");
      expect(lastStep.type).toBe("UI");
      return body;
    }

    await new Promise((resolve) => setTimeout(resolve, 100));
  }
}

async function untilFlowFinished({ name, id, token }) {
  const startTime = Date.now();
  const timeout = 1000; // 1 seconds timeout on polling

  while (true) {
    if (Date.now() - startTime > timeout) {
      throw new Error(
        `timed out waiting for flow run to reach a completed state after ${timeout}ms`
      );
    }

    const { status, body } = await getFlowRun({ name, id, token });
    expect(status).toEqual(200);

    if (body.status === "COMPLETED" || body.status === "FAILED") {
      for (const step of body.steps) {
        // Steps can only be COMPLETED or FAILED when flow has finished
        expect(step.status === "COMPLETED" || step.status === "FAILED").toBe(
          true
        );
      }
      return body;
    }

    await new Promise((resolve) => setTimeout(resolve, 100));
  }
}
