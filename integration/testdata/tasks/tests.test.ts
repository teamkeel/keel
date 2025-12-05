import { resetDatabase, models, flows } from "@teamkeel/testing";
import { useDatabase } from "@teamkeel/sdk";
import { beforeEach, expect, test } from "vitest";

beforeEach(resetDatabase);

test("tasks - create", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({
    topic: "DispatchOrder",
    body: {
      data: {
        orderDate: new Date(2025, 6, 15),
        shipByDate: new Date(2025, 6, 30),
      },
    },
    token: token,
  });
  expect(resCreate.status).toBe(200);
  expect(resCreate.body).toEqual({
    createdAt: expect.any(String),
    id: expect.any(String),
    name: "DispatchOrder",
    status: "NEW",
    updatedAt: expect.any(String),
  });

  const res = await getTaskQueue({ topic: "DispatchOrder", token: token });
  expect(res.status).toBe(200);
  expect(res.body).toEqual([
    {
      createdAt: expect.any(String),
      id: expect.any(String),
      name: "DispatchOrder",
      status: "NEW",
      updatedAt: expect.any(String),
    },
  ]);

  const tasks = await useDatabase()
    .selectFrom("dispatch_order")
    .selectAll()
    .execute();

  expect(tasks).toEqual([
    {
      id: expect.any(String),
      orderDate: expect.any(Date),
      shipByDate: expect.any(Date),
      createdAt: expect.any(Date),
      updatedAt: expect.any(Date),
      keelTaskId: resCreate.body.id,
    },
  ]);
});

test("tasks - create - no fields", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({
    topic: "NoFields",
    body: {},
    token: token,
  });
  expect(resCreate.status).toBe(200);
  expect(resCreate.body).toEqual({
    createdAt: expect.any(String),
    id: expect.any(String),
    name: "NoFields",
    status: "NEW",
    updatedAt: expect.any(String),
  });

  const res = await getTaskQueue({ topic: "NoFields", token: token });
  expect(res.status).toBe(200);
  expect(res.body).toEqual([
    {
      createdAt: expect.any(String),
      id: expect.any(String),
      name: "NoFields",
      status: "NEW",
      updatedAt: expect.any(String),
    },
  ]);
});

test("tasks - start", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({
    topic: "NoFields",
    body: {},
    token: token,
  });
  expect(resCreate.status).toBe(200);
  expect(resCreate.body).toEqual({
    createdAt: expect.any(String),
    id: expect.any(String),
    name: "NoFields",
    status: "NEW",
    updatedAt: expect.any(String),
  });

  const taskId = resCreate.body.id;

  const res = await startTask({ topic: "NoFields", token: token, id: taskId });
  expect(res.status).toBe(200);
  expect(res.body).toEqual({
    createdAt: expect.any(String),
    id: expect.any(String),
    name: "NoFields",
    status: "NEW",
    updatedAt: expect.any(String),
    flowRunId: expect.any(String),
  });

  const flowRunId = res.body.flowRunId;
  const flow = await flows.noFields
    .withAuthToken(token)
    .untilFinished(flowRunId);

  expect(flow).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "COMPLETED",
    name: "NoFields",
    startedBy: expect.any(String),
    input: {
      entityId: expect.any(String),
    },
    error: null,
    data: null,
    config: {
      title: "No fields",
    },
    steps: [
      {
        id: expect.any(String),
        name: "return task entity id",
        runId: expect.any(String),
        status: "COMPLETED",
        type: "FUNCTION",
        value: expect.any(String),
        error: null,
        stage: null,
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        ui: null,
      },
    ],
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
  });
});

test("tasks - list", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const t1 = await createTask({
    topic: "DispatchOrder",
    body: {
      data: {
        orderDate: new Date(2025, 6, 9),
        shipByDate: new Date(2025, 6, 20),
      },
    },
    token: token,
  });
  expect(t1.status).toBe(200);
  const t2 = await createTask({
    topic: "DispatchOrder",
    body: {
      data: {
        orderDate: new Date(2025, 6, 15),
        shipByDate: new Date(2025, 6, 30),
      },
    },
    token: token,
  });
  expect(t2.status).toBe(200);
  const t3 = await createTask({
    topic: "DispatchOrder",
    body: {
      data: {
        orderDate: new Date(2025, 6, 14),
        shipByDate: new Date(2025, 6, 30),
      },
    },
    token: token,
  });
  expect(t3.status).toBe(200);
  const t4 = await createTask({
    topic: "DispatchOrder",
    body: {
      data: {
        orderDate: new Date(2025, 6, 10),
        shipByDate: new Date(2025, 6, 20),
      },
    },
    token: token,
  });
  expect(t4.status).toBe(200);
  const t5 = await createTask({
    topic: "DispatchOrder",
    body: {
      data: {
        orderDate: new Date(2025, 6, 10),
        shipByDate: new Date(2025, 6, 20),
      },
    },
    token: token,
  });
  expect(t5.status).toBe(200);

  const res = await getTaskQueue({ topic: "DispatchOrder", token: token });
  expect(res.status).toBe(200);
  expect(res.body).toEqual([t1.body, t5.body, t4.body, t3.body, t2.body]);
});

test("tasks - next - no tasks exist", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  const res = await nextTask({ topic: "DispatchOrder", token: token });
  expect(res.status).toBe(404);
  expect(res.body).toEqual({
    code: "ERR_RECORD_NOT_FOUND",
    message: "Not found",
  });
});

test("tasks - next - successfully assigned", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({
    topic: "DispatchOrder",
    body: {
      data: {
        orderDate: new Date(2025, 6, 9),
        shipByDate: new Date(2025, 6, 20),
      },
    },
    token: token,
  });
  expect(resCreate.status).toBe(200);

  const res = await nextTask({ topic: "DispatchOrder", token: token });
  expect(res.status).toBe(200);
  expect(res.body).toEqual({
    createdAt: expect.any(String),
    id: resCreate.body.id,
    name: "DispatchOrder",
    status: "ASSIGNED",
    updatedAt: expect.any(String),
    assignedTo: expect.any(String),
    assignedAt: expect.any(String),
  });
});

test("tasks - next - already assigned", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate1 = await createTask({
    topic: "DispatchOrder",
    body: {
      data: {
        orderDate: new Date(2025, 6, 9),
        shipByDate: new Date(2025, 4, 20),
      },
    },
    token: token,
  });
  expect(resCreate1.status).toBe(200);

  const resCreate2 = await createTask({
    topic: "DispatchOrder",
    body: {
      data: {
        orderDate: new Date(2025, 6, 9),
        shipByDate: new Date(2025, 6, 20),
      },
    },
    token: token,
  });
  expect(resCreate2.status).toBe(200);

  const resNext = await nextTask({ topic: "DispatchOrder", token: token });
  expect(resNext.status).toBe(200);
  expect(resNext.body).toEqual({
    createdAt: expect.any(String),
    id: resCreate1.body.id,
    name: "DispatchOrder",
    status: "ASSIGNED",
    updatedAt: expect.any(String),
    assignedTo: expect.any(String),
    assignedAt: expect.any(String),
  });

  const resNextAgain = await nextTask({ topic: "DispatchOrder", token: token });
  expect(resNextAgain.status).toBe(200);
  expect(resNextAgain.body).toEqual(resNext.body);

  const resList = await getTaskQueue({ topic: "DispatchOrder", token: token });
  expect(resList.status).toBe(200);
  expect(resList.body).toEqual([resNext.body, resCreate2.body]);
});

test("tasks - assign - successfully assigned", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({
    topic: "DispatchOrder",
    body: {
      data: {
        orderDate: new Date(2025, 6, 9),
        shipByDate: new Date(2025, 6, 20),
      },
    },
    token: token,
  });
  expect(resCreate.status).toBe(200);

  const identity = await models.identity.findOne({
    email: "admin@keel.xyz",
    issuer: "https://keel.so",
  });

  const res = await assignTask({
    topic: "DispatchOrder",
    token: token,
    id: resCreate.body.id,
    body: { assigned_to: identity!.id },
  });

  expect(res.status).toBe(200);
  expect(res.body).toEqual({
    createdAt: expect.any(String),
    id: resCreate.body.id,
    name: "DispatchOrder",
    status: "ASSIGNED",
    updatedAt: expect.any(String),
    assignedTo: identity!.id,
    assignedAt: expect.any(String),
  });
});

test("tasks - assign - task not found", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const identity = await models.identity.findOne({
    email: "admin@keel.xyz",
    issuer: "https://keel.so",
  });

  const res = await assignTask({
    topic: "DispatchOrder",
    token: token,
    id: "non-existent-id",
    body: { assigned_to: identity!.id },
  });

  expect(res.status).toBe(404);
  expect(res.body).toEqual({
    code: "ERR_RECORD_NOT_FOUND",
    message: "Not found",
  });
});

test("tasks - assign - completed task cannot be assigned", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({
    topic: "DispatchOrder",
    body: {
      data: {
        orderDate: new Date(2025, 6, 9),
        shipByDate: new Date(2025, 6, 20),
      },
    },
    token: token,
  });
  expect(resCreate.status).toBe(200);

  const resComplete = await completeTask({
    topic: "DispatchOrder",
    token: token,
    id: resCreate.body.id,
  });
  expect(resComplete.status).toBe(200);
  expect(resComplete.body.status).toBe("COMPLETED");

  const identity = await models.identity.findOne({
    email: "admin@keel.xyz",
    issuer: "https://keel.so",
  });

  const res = await assignTask({
    topic: "DispatchOrder",
    token: token,
    id: resCreate.body.id,
    body: { assigned_to: identity!.id },
  });

  expect(res.status).toBe(500);
  expect(res.body).toEqual({
    code: "ERR_INTERNAL",
    message: "error executing request (task already completed)",
  });
});

test("tasks - assign - reassign to different user", async () => {
  const tokenAdmin = await getToken({ email: "admin@keel.xyz" });
  const tokenOther = await getToken({ email: "other@keel.xyz" });

  const resCreate = await createTask({
    topic: "DispatchOrder",
    body: {
      data: {
        orderDate: new Date(2025, 6, 9),
        shipByDate: new Date(2025, 6, 20),
      },
    },
    token: tokenAdmin,
  });
  expect(resCreate.status).toBe(200);

  const adminIdentity = await models.identity.findOne({
    email: "admin@keel.xyz",
    issuer: "https://keel.so",
  });
  const otherIdentity = await models.identity.findOne({
    email: "other@keel.xyz",
    issuer: "https://keel.so",
  });

  const resAssign1 = await assignTask({
    topic: "DispatchOrder",
    token: tokenAdmin,
    id: resCreate.body.id,
    body: { assigned_to: adminIdentity!.id },
  });
  expect(resAssign1.status).toBe(200);
  expect(resAssign1.body.assignedTo).toBe(adminIdentity!.id);

  const resAssign2 = await assignTask({
    topic: "DispatchOrder",
    token: tokenAdmin,
    id: resCreate.body.id,
    body: { assigned_to: otherIdentity!.id },
  });
  expect(resAssign2.status).toBe(200);
  expect(resAssign2.body).toEqual({
    createdAt: expect.any(String),
    id: resCreate.body.id,
    name: "DispatchOrder",
    status: "ASSIGNED",
    updatedAt: expect.any(String),
    assignedTo: otherIdentity!.id,
    assignedAt: expect.any(String),
  });
});

test("tasks - assign - missing assigned_to in body", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({
    topic: "DispatchOrder",
    body: {
      data: {
        orderDate: new Date(2025, 6, 9),
        shipByDate: new Date(2025, 6, 20),
      },
    },
    token: token,
  });
  expect(resCreate.status).toBe(200);

  const res = await assignTask({
    topic: "DispatchOrder",
    token: token,
    id: resCreate.body.id,
    body: {},
  });

  expect(res.status).toBe(400);
  expect(res.body).toEqual({
    code: "ERR_INPUT_MALFORMED",
    message: "data not correctly formatted",
  });
});

test("tasks - defer - successfully deferred", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({
    topic: "DispatchOrder",
    body: {
      data: {
        orderDate: new Date(2025, 6, 9),
        shipByDate: new Date(2025, 6, 20),
      },
    },
    token: token,
  });
  expect(resCreate.status).toBe(200);

  const deferUntil = new Date(2025, 7, 15).toISOString();
  const res = await deferTask({
    topic: "DispatchOrder",
    token: token,
    id: resCreate.body.id,
    body: { defer_until: deferUntil },
  });

  expect(res.status).toBe(200);
  expect(res.body).toEqual({
    createdAt: expect.any(String),
    id: resCreate.body.id,
    name: "DispatchOrder",
    status: "DEFERRED",
    updatedAt: expect.any(String),
    deferredUntil: expect.any(String),
  });
});

test("tasks - defer - task not found", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const deferUntil = new Date(2025, 7, 15).toISOString();
  const res = await deferTask({
    topic: "DispatchOrder",
    token: token,
    id: "non-existent-id",
    body: { defer_until: deferUntil },
  });

  expect(res.status).toBe(404);
  expect(res.body).toEqual({
    code: "ERR_RECORD_NOT_FOUND",
    message: "Not found",
  });
});

test("tasks - defer - completed task cannot be deferred", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({
    topic: "DispatchOrder",
    body: {
      data: {
        orderDate: new Date(2025, 6, 9),
        shipByDate: new Date(2025, 6, 20),
      },
    },
    token: token,
  });
  expect(resCreate.status).toBe(200);

  const resComplete = await completeTask({
    topic: "DispatchOrder",
    token: token,
    id: resCreate.body.id,
  });
  expect(resComplete.status).toBe(200);
  expect(resComplete.body.status).toBe("COMPLETED");

  const deferUntil = new Date(2025, 7, 15).toISOString();
  const res = await deferTask({
    topic: "DispatchOrder",
    token: token,
    id: resCreate.body.id,
    body: { defer_until: deferUntil },
  });

  expect(res.status).toBe(500);
  expect(res.body).toEqual({
    code: "ERR_INTERNAL",
    message: "error executing request (task already completed)",
  });
});

test("tasks - defer - missing defer_until in body", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({
    topic: "DispatchOrder",
    body: {
      data: {
        orderDate: new Date(2025, 6, 9),
        shipByDate: new Date(2025, 6, 20),
      },
    },
    token: token,
  });
  expect(resCreate.status).toBe(200);

  const res = await deferTask({
    topic: "DispatchOrder",
    token: token,
    id: resCreate.body.id,
    body: {},
  });

  expect(res.status).toBe(400);
  expect(res.body).toEqual({
    code: "ERR_INPUT_MALFORMED",
    message: "data not correctly formatted",
  });
});

test("tasks - defer - invalid defer_until format", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({
    topic: "DispatchOrder",
    body: {
      data: {
        orderDate: new Date(2025, 6, 9),
        shipByDate: new Date(2025, 6, 20),
      },
    },
    token: token,
  });
  expect(resCreate.status).toBe(200);

  const res = await deferTask({
    topic: "DispatchOrder",
    token: token,
    id: resCreate.body.id,
    body: { defer_until: "not-a-valid-date" },
  });

  expect(res.status).toBe(400);
  expect(res.body).toEqual({
    code: "ERR_INPUT_MALFORMED",
    message: "date not correctly formatted",
  });
});

test("tasks - defer - deferred task not assigned via next", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({
    topic: "DispatchOrder",
    body: {
      data: {
        orderDate: new Date(2025, 6, 9),
        shipByDate: new Date(2025, 6, 20),
      },
    },
    token: token,
  });
  expect(resCreate.status).toBe(200);

  const futureDate = new Date();
  futureDate.setDate(futureDate.getDate() + 7);
  const res = await deferTask({
    topic: "DispatchOrder",
    token: token,
    id: resCreate.body.id,
    body: { defer_until: futureDate.toISOString() },
  });
  expect(res.status).toBe(200);
  expect(res.body.status).toBe("DEFERRED");

  const resNext = await nextTask({ topic: "DispatchOrder", token: token });
  expect(resNext.status).toBe(404);
  expect(resNext.body).toEqual({
    code: "ERR_RECORD_NOT_FOUND",
    message: "Not found",
  });
});

test("tasks - defer - non-deferred task picked over deferred task", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate1 = await createTask({
    topic: "DispatchOrder",
    body: {
      data: {
        orderDate: new Date(2025, 6, 9),
        shipByDate: new Date(2025, 6, 15),
      },
    },
    token: token,
  });
  expect(resCreate1.status).toBe(200);

  const futureDate = new Date();
  futureDate.setDate(futureDate.getDate() + 7);
  const resDefer = await deferTask({
    topic: "DispatchOrder",
    token: token,
    id: resCreate1.body.id,
    body: { defer_until: futureDate.toISOString() },
  });
  expect(resDefer.status).toBe(200);

  const resCreate2 = await createTask({
    topic: "DispatchOrder",
    body: {
      data: {
        orderDate: new Date(2025, 6, 10),
        shipByDate: new Date(2025, 6, 20),
      },
    },
    token: token,
  });
  expect(resCreate2.status).toBe(200);

  const resNext = await nextTask({ topic: "DispatchOrder", token: token });
  expect(resNext.status).toBe(200);
  expect(resNext.body.id).toBe(resCreate2.body.id);
  expect(resNext.body.status).toBe("ASSIGNED");
});

test("tasks - defer - deferred task assigned after defer_until passes", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({
    topic: "DispatchOrder",
    body: {
      data: {
        orderDate: new Date(2025, 6, 9),
        shipByDate: new Date(2025, 6, 20),
      },
    },
    token: token,
  });
  expect(resCreate.status).toBe(200);

  const pastDate = new Date();
  pastDate.setDate(pastDate.getDate() - 1);
  const resDefer = await deferTask({
    topic: "DispatchOrder",
    token: token,
    id: resCreate.body.id,
    body: { defer_until: pastDate.toISOString() },
  });
  expect(resDefer.status).toBe(200);
  expect(resDefer.body.status).toBe("DEFERRED");

  const resNext = await nextTask({ topic: "DispatchOrder", token: token });
  expect(resNext.status).toBe(200);
  expect(resNext.body.id).toBe(resCreate.body.id);
  expect(resNext.body.status).toBe("ASSIGNED");
});

test("tasks - cancel - successfully cancelled", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({
    topic: "DispatchOrder",
    body: {
      data: {
        orderDate: new Date(2025, 6, 9),
        shipByDate: new Date(2025, 6, 20),
      },
    },
    token: token,
  });
  expect(resCreate.status).toBe(200);

  const res = await cancelTask({
    topic: "DispatchOrder",
    token: token,
    id: resCreate.body.id,
  });

  expect(res.status).toBe(200);
  expect(res.body).toEqual({
    createdAt: expect.any(String),
    id: resCreate.body.id,
    name: "DispatchOrder",
    status: "CANCELLED",
    updatedAt: expect.any(String),
    resolvedAt: expect.any(String),
  });
});

test("tasks - cancel - task not found", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const res = await cancelTask({
    topic: "DispatchOrder",
    token: token,
    id: "non-existent-id",
  });

  expect(res.status).toBe(404);
  expect(res.body).toEqual({
    code: "ERR_RECORD_NOT_FOUND",
    message: "Not found",
  });
});

test("tasks - cancel - completed task cannot be cancelled", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({
    topic: "DispatchOrder",
    body: {
      data: {
        orderDate: new Date(2025, 6, 9),
        shipByDate: new Date(2025, 6, 20),
      },
    },
    token: token,
  });
  expect(resCreate.status).toBe(200);

  const resComplete = await completeTask({
    topic: "DispatchOrder",
    token: token,
    id: resCreate.body.id,
  });
  expect(resComplete.status).toBe(200);
  expect(resComplete.body.status).toBe("COMPLETED");

  const res = await cancelTask({
    topic: "DispatchOrder",
    token: token,
    id: resCreate.body.id,
  });

  expect(res.status).toBe(500);
  expect(res.body).toEqual({
    code: "ERR_INTERNAL",
    message: "error executing request (task already completed)",
  });
});

test("tasks - cancel - cancelled task not assigned via next", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({
    topic: "DispatchOrder",
    body: {
      data: {
        orderDate: new Date(2025, 6, 9),
        shipByDate: new Date(2025, 6, 20),
      },
    },
    token: token,
  });
  expect(resCreate.status).toBe(200);

  const resCancel = await cancelTask({
    topic: "DispatchOrder",
    token: token,
    id: resCreate.body.id,
  });
  expect(resCancel.status).toBe(200);
  expect(resCancel.body.status).toBe("CANCELLED");

  const resNext = await nextTask({ topic: "DispatchOrder", token: token });
  expect(resNext.status).toBe(404);
  expect(resNext.body).toEqual({
    code: "ERR_RECORD_NOT_FOUND",
    message: "Not found",
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

async function createTask({ topic, body, token }) {
  const url = `${process.env.KEEL_TESTING_API_URL}/topics/json/${topic}/tasks`;
  const res = await fetch(url, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: "Bearer " + token,
    },
    body: JSON.stringify(body),
  });

  return {
    status: res.status,
    body: await res.json(),
  };
}

async function getTaskQueue({ topic, token }) {
  const url = `${process.env.KEEL_TESTING_API_URL}/topics/json/${topic}/tasks`;

  const res = await fetch(url, {
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

async function nextTask({ topic, token }) {
  const url = `${process.env.KEEL_TESTING_API_URL}/topics/json/${topic}/tasks/next`;
  const res = await fetch(url, {
    method: "POST",
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

async function startTask({ topic, token, id }) {
  const url = `${process.env.KEEL_TESTING_API_URL}/topics/json/${topic}/tasks/${id}/start`;
  const res = await fetch(url, {
    method: "PUT",
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

async function assignTask({ topic, token, id, body }) {
  const url = `${process.env.KEEL_TESTING_API_URL}/topics/json/${topic}/tasks/${id}/assign`;
  const res = await fetch(url, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
      Authorization: "Bearer " + token,
    },
    body: JSON.stringify(body),
  });

  return {
    status: res.status,
    body: await res.json(),
  };
}

async function completeTask({ topic, token, id }) {
  const url = `${process.env.KEEL_TESTING_API_URL}/topics/json/${topic}/tasks/${id}/complete`;
  const res = await fetch(url, {
    method: "PUT",
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

async function deferTask({ topic, token, id, body }) {
  const url = `${process.env.KEEL_TESTING_API_URL}/topics/json/${topic}/tasks/${id}/defer`;
  const res = await fetch(url, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
      Authorization: "Bearer " + token,
    },
    body: JSON.stringify(body),
  });

  return {
    status: res.status,
    body: await res.json(),
  };
}

async function cancelTask({ topic, token, id }) {
  const url = `${process.env.KEEL_TESTING_API_URL}/topics/json/${topic}/tasks/${id}/cancel`;
  const res = await fetch(url, {
    method: "PUT",
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
