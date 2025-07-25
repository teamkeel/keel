import { resetDatabase, models } from "@teamkeel/testing";
import { beforeEach, expect, test } from "vitest";

beforeEach(resetDatabase);

test("flows - error in step with retries", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  const res = await startFlow({ name: "ErrorInStep", token, body: {} });
  expect(res.status).toBe(200);
  expect(res.body).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "RUNNING",
    startedBy: expect.any(String),
    name: "ErrorInStep",
    input: {},
    data: null,
    steps: [
      {
        id: expect.any(String),
        name: "erroring step",
        runId: expect.any(String),
        stage: null,
        status: "NEW",
        type: "FUNCTION",
        value: null,
        error: null,
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        startTime: null,
        endTime: null,
        ui: null,
      },
    ],
    createdAt: expect.any(String),
    updatedAt: expect.any(String),
    config: {
      title: "Error in step",
    },
  });

  const flow = await untilFlowFinished({
    name: "ErrorInStep",
    id: res.body.id,
    token,
  });

  // We are expecting 3 steps (the initial step + 2 retries)
  expect(flow).toEqual({
    id: res.body.id,
    traceId: res.body.traceId,
    status: "FAILED",
    name: "ErrorInStep",
    startedBy: expect.any(String),
    input: {},
    data: null,
    steps: [
      {
        id: res.body.steps[0].id,
        name: "erroring step",
        runId: res.body.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "Error in step",
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        startTime: expect.any(String),
        endTime: expect.any(String),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "erroring step",
        runId: res.body.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "Error in step",
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        startTime: expect.any(String),
        endTime: expect.any(String),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "erroring step",
        runId: res.body.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "Error in step",
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        startTime: expect.any(String),
        endTime: expect.any(String),
        ui: null,
      },
    ],
    createdAt: res.body.createdAt,
    updatedAt: expect.any(String),
    config: {
      title: "Error in step",
    },
  });
});

test("flows - on failure callback", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  const res = await startFlow({ name: "OnFailureCallback", token, body: {} });
  expect(res.status).toBe(200);
  expect(res.body).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "RUNNING",
    startedBy: expect.any(String),
    name: "OnFailureCallback",
    input: {},
    data: null,
    steps: [
      {
        id: expect.any(String),
        name: "erroring step",
        runId: expect.any(String),
        stage: null,
        status: "NEW",
        type: "FUNCTION",
        value: null,
        error: null,
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        startTime: null,
        endTime: null,
        ui: null,
      },
    ],
    createdAt: expect.any(String),
    updatedAt: expect.any(String),
    config: {
      title: "On failure callback",
    },
  });

  const flow = await untilFlowFinished({
    name: "OnFailureCallback",
    id: res.body.id,
    token,
  });

  // We are expecting 2 steps (the initial step + 1 retry)
  expect(flow).toEqual({
    id: res.body.id,
    traceId: res.body.traceId,
    status: "FAILED",
    name: "OnFailureCallback",
    startedBy: expect.any(String),
    input: {},
    data: null,
    steps: [
      {
        id: res.body.steps[0].id,
        name: "erroring step",
        runId: res.body.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "1 exists",
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        startTime: expect.any(String),
        endTime: expect.any(String),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "erroring step",
        runId: res.body.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "2 exists",
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        startTime: expect.any(String),
        endTime: expect.any(String),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "erroring step",
        runId: res.body.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "3 exists",
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        startTime: expect.any(String),
        endTime: expect.any(String),
        ui: null,
      },
    ],
    createdAt: res.body.createdAt,
    updatedAt: expect.any(String),
    config: {
      title: "On failure callback",
    },
  });

  expect(await models.thing.findMany()).toHaveLength(0);
});

test("flows - do not retry", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  const res = await startFlow({ name: "DoNotRetry", token, body: {} });
  expect(res.status).toBe(200);
  expect(res.body).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "RUNNING",
    startedBy: expect.any(String),
    name: "DoNotRetry",
    input: {},
    data: null,
    steps: [
      {
        id: expect.any(String),
        name: "erroring step",
        runId: expect.any(String),
        stage: null,
        status: "NEW",
        type: "FUNCTION",
        value: null,
        error: null,
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        startTime: null,
        endTime: null,
        ui: null,
      },
    ],
    createdAt: expect.any(String),
    updatedAt: expect.any(String),
    config: {
      title: "Do not retry",
    },
  });

  const flow = await untilFlowFinished({
    name: "DoNotRetry",
    id: res.body.id,
    token,
  });

  // We are expecting 2 steps (the initial step + 1 retry)
  expect(flow).toEqual({
    id: res.body.id,
    traceId: res.body.traceId,
    status: "FAILED",
    name: "DoNotRetry",
    startedBy: expect.any(String),
    input: {},
    data: null,
    steps: [
      {
        id: res.body.steps[0].id,
        name: "erroring step",
        runId: res.body.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "do not retry!",
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        startTime: expect.any(String),
        endTime: expect.any(String),
        ui: null,
      },
    ],
    createdAt: res.body.createdAt,
    updatedAt: expect.any(String),
    config: {
      title: "Do not retry",
    },
  });

  expect(await models.thing.findMany()).toHaveLength(0);
});

test("flows - eventual step success", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  const res = await startFlow({ name: "EventualStepSuccess", token, body: {} });
  expect(res.status).toBe(200);
  expect(res.body).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "RUNNING",
    startedBy: expect.any(String),
    name: "EventualStepSuccess",
    input: {},
    data: null,
    steps: [
      {
        id: expect.any(String),
        name: "erroring step",
        runId: expect.any(String),
        stage: null,
        status: "NEW",
        type: "FUNCTION",
        value: null,
        error: null,
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        startTime: null,
        endTime: null,
        ui: null,
      },
    ],
    createdAt: expect.any(String),
    updatedAt: expect.any(String),
    config: {
      title: "Eventual step success",
    },
  });

  const flow = await untilFlowFinished({
    name: "EventualStepSuccess",
    id: res.body.id,
    token,
  });

  expect(flow).toEqual({
    id: res.body.id,
    traceId: res.body.traceId,
    status: "COMPLETED",
    name: "EventualStepSuccess",
    startedBy: expect.any(String),
    input: {},
    data: null,
    steps: [
      {
        id: res.body.steps[0].id,
        name: "erroring step",
        runId: res.body.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "Error at attempt 1 of 4",
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        startTime: expect.any(String),
        endTime: expect.any(String),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "erroring step",
        runId: res.body.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "Error at attempt 2 of 4",
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        startTime: expect.any(String),
        endTime: expect.any(String),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "erroring step",
        runId: res.body.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "Error at attempt 3 of 4",
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        startTime: expect.any(String),
        endTime: expect.any(String),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "erroring step",
        runId: res.body.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "Error at attempt 4 of 4",
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        startTime: expect.any(String),
        endTime: expect.any(String),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "erroring step",
        runId: res.body.id,
        stage: null,
        status: "COMPLETED",
        type: "FUNCTION",
        value: "Success at attempt 5",
        error: null,
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        startTime: expect.any(String),
        endTime: expect.any(String),
        ui: null,
      },
    ],
    createdAt: res.body.createdAt,
    updatedAt: expect.any(String),
    config: {
      title: "Eventual step success",
    },
  });
});

test("flows - error in flow", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  const res = await startFlow({ name: "ErrorInFlow", token, body: {} });
  expect(res.status).toBe(200);
  expect(res.body).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "FAILED",
    name: "ErrorInFlow",
    startedBy: expect.any(String),
    input: {},
    data: null,
    steps: [],
    createdAt: expect.any(String),
    updatedAt: expect.any(String),
    config: {
      title: "Error in flow",
    },
  });
});

test("flows - timeout step", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  const res = await startFlow({ name: "TimeoutStep", token, body: {} });
  expect(res.status).toBe(200);
  expect(res.body).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "RUNNING",
    name: "TimeoutStep",
    startedBy: expect.any(String),
    input: {},
    data: null,
    steps: [
      {
        id: expect.any(String),
        name: "timeout step",
        runId: expect.any(String),
        stage: null,
        status: "NEW",
        type: "FUNCTION",
        value: null,
        error: null,
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        ui: null,
        startTime: null,
        endTime: null,
      },
    ],
    createdAt: expect.any(String),
    updatedAt: expect.any(String),
    config: {
      title: "Timeout step",
    },
  });

  const flow = await untilFlowFinished({
    name: "TimeoutStep",
    id: res.body.id,
    token,
  });

  // We are expecting 5 steps (the default)
  expect(flow).toEqual({
    id: res.body.id,
    traceId: res.body.traceId,
    status: "FAILED",
    name: "TimeoutStep",
    startedBy: expect.any(String),
    input: {},
    data: null,
    steps: [
      {
        id: res.body.steps[0].id,
        name: "timeout step",
        runId: res.body.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "Step function timed out after 10ms",
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        startTime: expect.any(String),
        endTime: expect.any(String),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "timeout step",
        runId: res.body.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "Step function timed out after 10ms",
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        startTime: expect.any(String),
        endTime: expect.any(String),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "timeout step",
        runId: res.body.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "Step function timed out after 10ms",
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        startTime: expect.any(String),
        endTime: expect.any(String),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "timeout step",
        runId: res.body.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "Step function timed out after 10ms",
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        startTime: expect.any(String),
        endTime: expect.any(String),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "timeout step",
        runId: res.body.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "Step function timed out after 10ms",
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        startTime: expect.any(String),
        endTime: expect.any(String),
        ui: null,
      },
    ],
    createdAt: res.body.createdAt,
    updatedAt: expect.any(String),
    config: {
      title: "Timeout step",
    },
  });
});

test("flows - error in validation", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  let { status, body } = await startFlow({
    name: "ErrorInValidation",
    token,
    body: {},
  });

  expect(status).toBe(200);
  expect(body.steps[0].status).toBe("PENDING");

  const runId = body.id;
  let stepId = body.steps[0].id;

  ({ status, body } = await putStepValues({
    name: "ErrorInValidation",
    runId,
    stepId,
    token,
    values: {},
    action: null,
  }));

  expect(status).toBe(200);
  expect(body).toEqual({
    createdAt: expect.any(String),
    data: null,
    id: expect.any(String),
    input: {},
    name: "ErrorInValidation",
    startedBy: expect.any(String),
    status: "FAILED",
    steps: [
      {
        createdAt: expect.any(String),
        endTime: expect.any(String),
        error: "something has gone wrong",
        id: expect.any(String),
        name: "first page",
        runId: expect.any(String),
        stage: null,
        startTime: expect.any(String),
        status: "FAILED",
        type: "UI",
        updatedAt: expect.any(String),
        value: null,
        ui: null,
      },
    ],
    traceId: expect.any(String),
    updatedAt: expect.any(String),
    config: {
      title: "Error in validation",
    },
  });
});

test("flows - duplicate step name", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  const res = await startFlow({ name: "DuplicateStepName", token, body: {} });
  expect(res.status).toBe(200);
  expect(res.body).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "RUNNING",
    name: "DuplicateStepName",
    startedBy: expect.any(String),
    input: {},
    data: null,
    steps: [
      {
        id: expect.any(String),
        name: "my step",
        runId: expect.any(String),
        stage: null,
        status: "NEW",
        type: "FUNCTION",
        value: null,
        error: null,
        ui: null,
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        endTime: null,
        startTime: null,
      },
    ],
    createdAt: expect.any(String),
    updatedAt: expect.any(String),
    config: {
      title: "Duplicate step name",
    },
  });

  const flow = await untilFlowFinished({
    name: "DuplicateStepName",
    id: res.body.id,
    token,
  });

  expect(flow).toEqual({
    id: res.body.id,
    traceId: res.body.traceId,
    status: "FAILED",
    name: "DuplicateStepName",
    startedBy: expect.any(String),
    input: {},
    data: null,
    steps: [
      {
        id: res.body.steps[0].id,
        name: "my step",
        runId: res.body.id,
        stage: null,
        status: "COMPLETED",
        type: "FUNCTION",
        value: null,
        error: null,
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        startTime: expect.any(String),
        endTime: expect.any(String),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "my step",
        runId: res.body.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "Duplicate step name: my step",
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        startTime: expect.any(String),
        endTime: expect.any(String),
        ui: null,
      },
    ],
    createdAt: res.body.createdAt,
    updatedAt: expect.any(String),
    config: {
      title: "Duplicate step name",
    },
  });
});

test("flows - duplicate step name and UI name", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  const res = await startFlow({
    name: "DuplicateStepUiName",
    token,
    body: {},
  });
  expect(res.status).toBe(200);
  expect(res.body).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "RUNNING",
    name: "DuplicateStepUiName",
    startedBy: expect.any(String),
    input: {},
    data: null,
    steps: [
      {
        id: expect.any(String),
        name: "my step",
        runId: expect.any(String),
        stage: null,
        status: "NEW",
        type: "FUNCTION",
        value: null,
        error: null,
        ui: null,
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        endTime: null,
        startTime: null,
      },
    ],
    createdAt: expect.any(String),
    updatedAt: expect.any(String),
    config: {
      title: "Duplicate step ui name",
    },
  });

  const flow = await untilFlowFinished({
    name: "DuplicateStepUiName",
    id: res.body.id,
    token,
  });

  expect(flow).toEqual({
    id: res.body.id,
    traceId: res.body.traceId,
    status: "FAILED",
    name: "DuplicateStepUiName",
    startedBy: expect.any(String),
    input: {},
    data: null,
    steps: [
      {
        id: res.body.steps[0].id,
        name: "my step",
        runId: res.body.id,
        stage: null,
        status: "COMPLETED",
        type: "FUNCTION",
        value: null,
        error: null,
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        startTime: expect.any(String),
        endTime: expect.any(String),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "my step",
        runId: res.body.id,
        stage: null,
        status: "FAILED",
        type: "UI",
        value: null,
        error: "Duplicate step name: my step",
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        startTime: expect.any(String),
        endTime: expect.any(String),
        ui: null,
      },
    ],
    createdAt: res.body.createdAt,
    updatedAt: expect.any(String),
    config: {
      title: "Duplicate step ui name",
    },
  });
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

async function putStepValues({ name, runId, stepId, values, token, action }) {
  let url = `${process.env.KEEL_TESTING_API_URL}/flows/json/${name}/${runId}/${stepId}`;
  if (action) {
    const queryString = new URLSearchParams({ action }).toString();
    url = `${url}?${queryString}`;
  }

  const res = await fetch(url, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
      Authorization: "Bearer " + token,
    },
    body: JSON.stringify(values),
  });

  return {
    status: res.status,
    body: await res.json(),
  };
}
