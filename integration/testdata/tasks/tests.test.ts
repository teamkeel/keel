import { resetDatabase, models, flows } from "@teamkeel/testing";
import { useDatabase } from "@teamkeel/sdk";
import { beforeEach, expect, test } from "vitest";

beforeEach(resetDatabase);

test("tasks - create", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({
    topic: "EmptyFlow",
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
    name: "EmptyFlow",
    status: "NEW",
    updatedAt: expect.any(String),
  });

  const res = await getTaskQueue({ topic: "EmptyFlow", token: token });
  expect(res.status).toBe(200);
  expect(res.body).toEqual([
    {
      createdAt: expect.any(String),
      id: expect.any(String),
      name: "EmptyFlow",
      status: "NEW",
      updatedAt: expect.any(String),
    },
  ]);
});

test("tasks - next assigns, start creates flow and completes task", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({
    topic: "EmptyFlow",
    body: {
      data: {
        orderDate: new Date(2025, 6, 9),
        shipByDate: new Date(2025, 6, 20),
      },
    },
    token: token,
  });
  expect(resCreate.status).toBe(200);

  const resNext = await nextTask({ topic: "EmptyFlow", token: token });
  expect(resNext.status).toBe(200);
  expect(resNext.body).toEqual({
    createdAt: expect.any(String),
    id: expect.any(String),
    name: "EmptyFlow",
    status: "ASSIGNED",
    updatedAt: expect.any(String),
    assignedTo: expect.any(String),
    assignedAt: expect.any(String),
  });

  const taskId = resNext.body.id;

  // No flow exists yet - flow is created when start is called
  expect(resNext.body.flowRunId).toBeUndefined();

  const resStart = await startTask({
    topic: "EmptyFlow",
    token: token,
    id: taskId,
  });
  expect(resStart.status).toBe(200);
  expect(resStart.body.status).toBe("STARTED");

  const flowRunId = resStart.body.flowRunId;
  expect(flowRunId).toBeDefined();

  // Wait for the flow to complete
  const completedFlow = await flows.emptyFlow
    .withAuthToken(token)
    .untilFinished(flowRunId);
  expect(completedFlow.status).toBe("COMPLETED");

  // Verify task was auto-completed when flow finished
  const taskFromDb = await (useDatabase() as any)
    .selectFrom("keel.task")
    .selectAll()
    .where("id", "=", taskId)
    .executeTakeFirst();
  expect(taskFromDb.status).toBe("COMPLETED");
  expect(taskFromDb.resolvedAt).not.toBeNull();

  const statusEntries = await (useDatabase() as any)
    .selectFrom("keel.task_status")
    .selectAll()
    .where("keel_task_id", "=", taskId)
    .orderBy("created_at", "asc")
    .execute();

  expect(statusEntries).toHaveLength(4);
  expect(statusEntries[0].status).toBe("NEW");
  expect(statusEntries[1].status).toBe("ASSIGNED");
  expect(statusEntries[2].status).toBe("STARTED");
  expect(statusEntries[2].flowRunId).toBe(flowRunId);
  expect(statusEntries[3].status).toBe("COMPLETED");
  expect(statusEntries[3].flowRunId).toBe(flowRunId);
});

test("tasks - list", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const t1 = await createTask({
    topic: "EmptyFlow",
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
    topic: "EmptyFlow",
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
    topic: "EmptyFlow",
    body: {
      data: {
        orderDate: new Date(2025, 6, 14),
        shipByDate: new Date(2025, 6, 30),
      },
    },
    token: token,
  });
  expect(t3.status).toBe(200);

  const res = await getTaskQueue({ topic: "EmptyFlow", token: token });
  expect(res.status).toBe(200);
  // Tasks ordered by shipByDate asc, then orderDate asc
  expect(res.body).toEqual([t1.body, t3.body, t2.body]);
});

test("tasks - next - no tasks exist", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  const res = await nextTask({ topic: "EmptyFlow", token: token });
  expect(res.status).toBe(404);
  expect(res.body).toEqual({
    code: "ERR_RECORD_NOT_FOUND",
    message: "Not found",
  });
});

test("tasks - next - successfully assigned", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({
    topic: "EmptyFlow",
    body: {
      data: {
        orderDate: new Date(2025, 6, 9),
        shipByDate: new Date(2025, 6, 20),
      },
    },
    token: token,
  });
  expect(resCreate.status).toBe(200);

  const res = await nextTask({ topic: "EmptyFlow", token: token });
  expect(res.status).toBe(200);
  expect(res.body).toEqual({
    createdAt: expect.any(String),
    id: resCreate.body.id,
    name: "EmptyFlow",
    status: "ASSIGNED",
    updatedAt: expect.any(String),
    assignedTo: expect.any(String),
    assignedAt: expect.any(String),
  });
});

test("tasks - next - already assigned", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  // Create task with earlier shipByDate - should be picked first due to @orderBy
  const resCreate1 = await createTask({
    topic: "EmptyFlow",
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
    topic: "EmptyFlow",
    body: {
      data: {
        orderDate: new Date(2025, 6, 9),
        shipByDate: new Date(2025, 6, 20),
      },
    },
    token: token,
  });
  expect(resCreate2.status).toBe(200);

  const resNext = await nextTask({ topic: "EmptyFlow", token: token });
  expect(resNext.status).toBe(200);
  expect(resNext.body).toEqual({
    createdAt: expect.any(String),
    id: resCreate1.body.id,
    name: "EmptyFlow",
    status: "ASSIGNED",
    updatedAt: expect.any(String),
    assignedTo: expect.any(String),
    assignedAt: expect.any(String),
  });

  // Calling next again returns the same already-assigned task
  const resNextAgain = await nextTask({ topic: "EmptyFlow", token: token });
  expect(resNextAgain.status).toBe(200);
  expect(resNextAgain.body.id).toBe(resNext.body.id);
});

test("tasks - assign - successfully assigned", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({
    topic: "EmptyFlow",
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
    topic: "EmptyFlow",
    token: token,
    id: resCreate.body.id,
    body: { assigned_to: identity!.id },
  });

  expect(res.status).toBe(200);
  expect(res.body).toEqual({
    createdAt: expect.any(String),
    id: resCreate.body.id,
    name: "EmptyFlow",
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
    topic: "EmptyFlow",
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
    topic: "EmptyFlow",
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
    topic: "EmptyFlow",
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
    topic: "EmptyFlow",
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
    topic: "EmptyFlow",
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
    topic: "EmptyFlow",
    token: tokenAdmin,
    id: resCreate.body.id,
    body: { assigned_to: adminIdentity!.id },
  });
  expect(resAssign1.status).toBe(200);
  expect(resAssign1.body.assignedTo).toBe(adminIdentity!.id);

  const resAssign2 = await assignTask({
    topic: "EmptyFlow",
    token: tokenAdmin,
    id: resCreate.body.id,
    body: { assigned_to: otherIdentity!.id },
  });
  expect(resAssign2.status).toBe(200);
  expect(resAssign2.body).toEqual({
    createdAt: expect.any(String),
    id: resCreate.body.id,
    name: "EmptyFlow",
    status: "ASSIGNED",
    updatedAt: expect.any(String),
    assignedTo: otherIdentity!.id,
    assignedAt: expect.any(String),
  });
});

test("tasks - assign - missing assigned_to in body", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({
    topic: "EmptyFlow",
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
    topic: "EmptyFlow",
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
    topic: "EmptyFlow",
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
    topic: "EmptyFlow",
    token: token,
    id: resCreate.body.id,
    body: { defer_until: deferUntil },
  });

  expect(res.status).toBe(200);
  expect(res.body).toEqual({
    createdAt: expect.any(String),
    id: resCreate.body.id,
    name: "EmptyFlow",
    status: "DEFERRED",
    updatedAt: expect.any(String),
    deferredUntil: expect.any(String),
  });
});

test("tasks - defer - task not found", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const deferUntil = new Date(2025, 7, 15).toISOString();
  const res = await deferTask({
    topic: "EmptyFlow",
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
    topic: "EmptyFlow",
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
    topic: "EmptyFlow",
    token: token,
    id: resCreate.body.id,
  });
  expect(resComplete.status).toBe(200);
  expect(resComplete.body.status).toBe("COMPLETED");

  const deferUntil = new Date(2025, 7, 15).toISOString();
  const res = await deferTask({
    topic: "EmptyFlow",
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
    topic: "EmptyFlow",
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
    topic: "EmptyFlow",
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
    topic: "EmptyFlow",
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
    topic: "EmptyFlow",
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
    topic: "EmptyFlow",
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
    topic: "EmptyFlow",
    token: token,
    id: resCreate.body.id,
    body: { defer_until: futureDate.toISOString() },
  });
  expect(res.status).toBe(200);
  expect(res.body.status).toBe("DEFERRED");

  const resNext = await nextTask({ topic: "EmptyFlow", token: token });
  expect(resNext.status).toBe(404);
  expect(resNext.body).toEqual({
    code: "ERR_RECORD_NOT_FOUND",
    message: "Not found",
  });
});

test("tasks - defer - non-deferred task picked over deferred task", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  // Task 1 would be first due to earlier shipByDate, but we'll defer it
  const resCreate1 = await createTask({
    topic: "EmptyFlow",
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
    topic: "EmptyFlow",
    token: token,
    id: resCreate1.body.id,
    body: { defer_until: futureDate.toISOString() },
  });
  expect(resDefer.status).toBe(200);

  // Task 2 has later shipByDate but should be picked since task 1 is deferred
  const resCreate2 = await createTask({
    topic: "EmptyFlow",
    body: {
      data: {
        orderDate: new Date(2025, 6, 10),
        shipByDate: new Date(2025, 6, 20),
      },
    },
    token: token,
  });
  expect(resCreate2.status).toBe(200);

  const resNext = await nextTask({ topic: "EmptyFlow", token: token });
  expect(resNext.status).toBe(200);
  expect(resNext.body.id).toBe(resCreate2.body.id);
  expect(resNext.body.status).toBe("ASSIGNED");
});

test("tasks - defer - deferred task assigned after defer_until passes", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({
    topic: "EmptyFlow",
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
    topic: "EmptyFlow",
    token: token,
    id: resCreate.body.id,
    body: { defer_until: pastDate.toISOString() },
  });
  expect(resDefer.status).toBe(200);
  expect(resDefer.body.status).toBe("DEFERRED");

  const resNext = await nextTask({ topic: "EmptyFlow", token: token });
  expect(resNext.status).toBe(200);
  expect(resNext.body.id).toBe(resCreate.body.id);
  expect(resNext.body.status).toBe("ASSIGNED");
});

test("tasks - cancel - successfully cancelled", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({
    topic: "EmptyFlow",
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
    topic: "EmptyFlow",
    token: token,
    id: resCreate.body.id,
  });

  expect(res.status).toBe(200);
  expect(res.body).toEqual({
    createdAt: expect.any(String),
    id: resCreate.body.id,
    name: "EmptyFlow",
    status: "CANCELLED",
    updatedAt: expect.any(String),
    resolvedAt: expect.any(String),
  });
});

test("tasks - cancel - task not found", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const res = await cancelTask({
    topic: "EmptyFlow",
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
    topic: "EmptyFlow",
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
    topic: "EmptyFlow",
    token: token,
    id: resCreate.body.id,
  });
  expect(resComplete.status).toBe(200);
  expect(resComplete.body.status).toBe("COMPLETED");

  const res = await cancelTask({
    topic: "EmptyFlow",
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
    topic: "EmptyFlow",
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
    topic: "EmptyFlow",
    token: token,
    id: resCreate.body.id,
  });
  expect(resCancel.status).toBe(200);
  expect(resCancel.body.status).toBe("CANCELLED");

  const resNext = await nextTask({ topic: "EmptyFlow", token: token });
  expect(resNext.status).toBe(404);
  expect(resNext.body).toEqual({
    code: "ERR_RECORD_NOT_FOUND",
    message: "Not found",
  });
});

test("tasks - unassign - successfully unassigned", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({
    topic: "EmptyFlow",
    body: {
      data: {
        orderDate: new Date(2025, 6, 9),
        shipByDate: new Date(2025, 6, 20),
      },
    },
    token: token,
  });
  expect(resCreate.status).toBe(200);

  // Assign the task first via next
  const resNext = await nextTask({ topic: "EmptyFlow", token: token });
  expect(resNext.status).toBe(200);
  expect(resNext.body.status).toBe("ASSIGNED");
  expect(resNext.body.assignedTo).toBeDefined();
  expect(resNext.body.assignedAt).toBeDefined();

  // Unassign the task
  const res = await unassignTask({
    topic: "EmptyFlow",
    token: token,
    id: resCreate.body.id,
  });

  expect(res.status).toBe(200);
  expect(res.body).toEqual({
    createdAt: expect.any(String),
    id: resCreate.body.id,
    name: "EmptyFlow",
    status: "NEW",
    updatedAt: expect.any(String),
  });
  // assignedTo and assignedAt should be cleared
  expect(res.body.assignedTo).toBeUndefined();
  expect(res.body.assignedAt).toBeUndefined();
});

test("tasks - unassign - task not found", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const res = await unassignTask({
    topic: "EmptyFlow",
    token: token,
    id: "non-existent-id",
  });

  expect(res.status).toBe(404);
  expect(res.body).toEqual({
    code: "ERR_RECORD_NOT_FOUND",
    message: "Not found",
  });
});

test("tasks - unassign - completed task cannot be unassigned", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({
    topic: "EmptyFlow",
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
    topic: "EmptyFlow",
    token: token,
    id: resCreate.body.id,
  });
  expect(resComplete.status).toBe(200);
  expect(resComplete.body.status).toBe("COMPLETED");

  const res = await unassignTask({
    topic: "EmptyFlow",
    token: token,
    id: resCreate.body.id,
  });

  expect(res.status).toBe(400);
  expect(res.body).toEqual({
    code: "ERR_INVALID_INPUT",
    message: "cannot unassign a completed or cancelled task",
  });
});

test("tasks - unassign - cancelled task cannot be unassigned", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({
    topic: "EmptyFlow",
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
    topic: "EmptyFlow",
    token: token,
    id: resCreate.body.id,
  });
  expect(resCancel.status).toBe(200);
  expect(resCancel.body.status).toBe("CANCELLED");

  const res = await unassignTask({
    topic: "EmptyFlow",
    token: token,
    id: resCreate.body.id,
  });

  expect(res.status).toBe(400);
  expect(res.body).toEqual({
    code: "ERR_INVALID_INPUT",
    message: "cannot unassign a completed or cancelled task",
  });
});

test("tasks - unassign - unassigned task available in queue", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({
    topic: "EmptyFlow",
    body: {
      data: {
        orderDate: new Date(2025, 6, 9),
        shipByDate: new Date(2025, 6, 20),
      },
    },
    token: token,
  });
  expect(resCreate.status).toBe(200);

  // Assign the task
  const resNext1 = await nextTask({ topic: "EmptyFlow", token: token });
  expect(resNext1.status).toBe(200);
  expect(resNext1.body.status).toBe("ASSIGNED");

  // Unassign the task
  const resUnassign = await unassignTask({
    topic: "EmptyFlow",
    token: token,
    id: resCreate.body.id,
  });
  expect(resUnassign.status).toBe(200);
  expect(resUnassign.body.status).toBe("NEW");

  // Task should now be available via next again
  const resNext2 = await nextTask({ topic: "EmptyFlow", token: token });
  expect(resNext2.status).toBe(200);
  expect(resNext2.body.id).toBe(resCreate.body.id);
  expect(resNext2.body.status).toBe("ASSIGNED");
});

test("tasks - unassign - creates NEW status entry", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({
    topic: "EmptyFlow",
    body: {
      data: {
        orderDate: new Date(2025, 6, 9),
        shipByDate: new Date(2025, 6, 20),
      },
    },
    token: token,
  });
  expect(resCreate.status).toBe(200);

  // Assign the task
  const resNext = await nextTask({ topic: "EmptyFlow", token: token });
  expect(resNext.status).toBe(200);

  const taskId = resNext.body.id;

  // Unassign the task
  const resUnassign = await unassignTask({
    topic: "EmptyFlow",
    token: token,
    id: taskId,
  });
  expect(resUnassign.status).toBe(200);

  // Verify status entries include the NEW status from unassign
  const statusEntries = await (useDatabase() as any)
    .selectFrom("keel.task_status")
    .selectAll()
    .where("keel_task_id", "=", taskId)
    .orderBy("created_at", "asc")
    .execute();

  expect(statusEntries).toHaveLength(3);
  expect(statusEntries[0].status).toBe("NEW");
  expect(statusEntries[1].status).toBe("ASSIGNED");
  expect(statusEntries[2].status).toBe("NEW"); // From unassign
});

test("tasks - unassign - can unassign NEW task (no-op)", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({
    topic: "EmptyFlow",
    body: {
      data: {
        orderDate: new Date(2025, 6, 9),
        shipByDate: new Date(2025, 6, 20),
      },
    },
    token: token,
  });
  expect(resCreate.status).toBe(200);
  expect(resCreate.body.status).toBe("NEW");

  // Unassign a task that was never assigned (should still work)
  const res = await unassignTask({
    topic: "EmptyFlow",
    token: token,
    id: resCreate.body.id,
  });

  expect(res.status).toBe(200);
  expect(res.body.status).toBe("NEW");
});

test("tasks - unassign - can unassign DEFERRED task", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({
    topic: "EmptyFlow",
    body: {
      data: {
        orderDate: new Date(2025, 6, 9),
        shipByDate: new Date(2025, 6, 20),
      },
    },
    token: token,
  });
  expect(resCreate.status).toBe(200);

  // Defer the task
  const futureDate = new Date();
  futureDate.setDate(futureDate.getDate() + 7);
  const resDefer = await deferTask({
    topic: "EmptyFlow",
    token: token,
    id: resCreate.body.id,
    body: { defer_until: futureDate.toISOString() },
  });
  expect(resDefer.status).toBe(200);
  expect(resDefer.body.status).toBe("DEFERRED");

  // Unassign the deferred task (should work and reset to NEW, but preserve deferredUntil)
  const res = await unassignTask({
    topic: "EmptyFlow",
    token: token,
    id: resCreate.body.id,
  });

  expect(res.status).toBe(200);
  expect(res.body.status).toBe("NEW");
  // deferredUntil is preserved - task still has the deferred_until set
  expect(res.body.deferredUntil).toBeDefined();
});

test("tasks - start - creates STARTED and COMPLETED status entries", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({
    topic: "EmptyFlow",
    body: {
      data: {
        orderDate: new Date(2025, 6, 9),
        shipByDate: new Date(2025, 6, 20),
      },
    },
    token: token,
  });
  expect(resCreate.status).toBe(200);

  const resNext = await nextTask({ topic: "EmptyFlow", token: token });
  expect(resNext.status).toBe(200);
  // No flow created during next - flow is created during start
  expect(resNext.body.flowRunId).toBeUndefined();

  const taskId = resNext.body.id;

  const resStart = await startTask({
    topic: "EmptyFlow",
    token: token,
    id: taskId,
  });
  expect(resStart.status).toBe(200);
  expect(resStart.body.status).toBe("STARTED");

  const flowRunId = resStart.body.flowRunId;
  expect(flowRunId).toBeDefined();

  // Wait for the flow to complete
  await flows.emptyFlow.withAuthToken(token).untilFinished(flowRunId);

  const statusEntries = await (useDatabase() as any)
    .selectFrom("keel.task_status")
    .selectAll()
    .where("keel_task_id", "=", taskId)
    .orderBy("created_at", "asc")
    .execute();

  expect(statusEntries).toHaveLength(4);
  expect(statusEntries[0].status).toBe("NEW");
  expect(statusEntries[0].flowRunId).toBeNull();
  expect(statusEntries[1].status).toBe("ASSIGNED");
  expect(statusEntries[1].flowRunId).toBeNull();
  expect(statusEntries[2].status).toBe("STARTED");
  expect(statusEntries[2].flowRunId).toBe(flowRunId);
  expect(statusEntries[3].status).toBe("COMPLETED");
  expect(statusEntries[3].flowRunId).toBe(flowRunId);
});

test("tasks - start - calling start on completed task returns error", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({
    topic: "EmptyFlow",
    body: {
      data: {
        orderDate: new Date(2025, 6, 9),
        shipByDate: new Date(2025, 6, 20),
      },
    },
    token: token,
  });
  expect(resCreate.status).toBe(200);

  const resNext = await nextTask({ topic: "EmptyFlow", token: token });
  expect(resNext.status).toBe(200);

  const resStart1 = await startTask({
    topic: "EmptyFlow",
    token: token,
    id: resNext.body.id,
  });
  expect(resStart1.status).toBe(200);
  expect(resStart1.body.status).toBe("STARTED");

  const flowRunId = resStart1.body.flowRunId;
  expect(flowRunId).toBeDefined();

  // Wait for the flow to complete (which auto-completes the task)
  await flows.emptyFlow.withAuthToken(token).untilFinished(flowRunId);

  // Calling start on a completed task returns an error
  const resStart2 = await startTask({
    topic: "EmptyFlow",
    token: token,
    id: resNext.body.id,
  });
  expect(resStart2.status).toBe(500);
  expect(resStart2.body).toEqual({
    code: "ERR_INTERNAL",
    message: "error executing request (task already completed)",
  });

  const taskId = resNext.body.id;
  const statusEntries = await (useDatabase() as any)
    .selectFrom("keel.task_status")
    .selectAll()
    .where("keel_task_id", "=", taskId)
    .orderBy("created_at", "asc")
    .execute();

  expect(statusEntries).toHaveLength(4);
  expect(statusEntries[0].status).toBe("NEW");
  expect(statusEntries[1].status).toBe("ASSIGNED");
  expect(statusEntries[2].status).toBe("STARTED");
  expect(statusEntries[3].status).toBe("COMPLETED");
});

test("tasks - flow completion auto-completes task", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({
    topic: "EmptyFlow",
    body: {
      data: {
        orderDate: new Date(2025, 6, 9),
        shipByDate: new Date(2025, 6, 20),
      },
    },
    token: token,
  });
  expect(resCreate.status).toBe(200);
  const taskId = resCreate.body.id;

  // Assign the task first
  const resNext = await nextTask({ topic: "EmptyFlow", token: token });
  expect(resNext.status).toBe(200);

  // Start the task - this creates and runs the flow
  const resStart = await startTask({
    topic: "EmptyFlow",
    token: token,
    id: taskId,
  });
  expect(resStart.status).toBe(200);
  expect(resStart.body.status).toBe("STARTED");

  const flowRunId = resStart.body.flowRunId;
  expect(flowRunId).toBeDefined();

  // Wait for the flow to complete
  const completedFlow = await flows.emptyFlow
    .withAuthToken(token)
    .untilFinished(flowRunId);
  expect(completedFlow.status).toBe("COMPLETED");

  // Fetch the task again to verify it was auto-completed when flow finished
  const taskFromDb = await (useDatabase() as any)
    .selectFrom("keel.task")
    .selectAll()
    .where("id", "=", taskId)
    .executeTakeFirst();

  expect(taskFromDb.status).toBe("COMPLETED");
  expect(taskFromDb.resolvedAt).not.toBeNull();
  expect(taskFromDb.flowRunId).toBe(flowRunId);

  // Verify task_status entries include COMPLETED with the flow run ID
  const statusEntries = await (useDatabase() as any)
    .selectFrom("keel.task_status")
    .selectAll()
    .where("keel_task_id", "=", taskId)
    .orderBy("created_at", "asc")
    .execute();

  expect(statusEntries).toHaveLength(4);
  expect(statusEntries[0].status).toBe("NEW");
  expect(statusEntries[0].flowRunId).toBeNull();

  expect(statusEntries[1].status).toBe("ASSIGNED");
  expect(statusEntries[1].flowRunId).toBeNull();

  expect(statusEntries[2].status).toBe("STARTED");
  expect(statusEntries[2].flowRunId).toBe(flowRunId);

  expect(statusEntries[3].status).toBe("COMPLETED");
  expect(statusEntries[3].flowRunId).toBe(flowRunId);
});

test("tasks - flow receives typed inputs from task fields", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  // Create test data with various Keel types
  const testDate = "2025-07-15";
  const testTimestamp = "2025-07-15T14:30:00.000Z";

  const resCreate = await createTask({
    topic: "InputsTask",
    body: {
      data: {
        textField: "hello world",
        numberField: 42,
        booleanField: true,
        dateField: testDate,
        timestampField: testTimestamp,
        decimalField: 123.456,
        enumField: "High",
        optionalTextField: "optional value",
      },
    },
    token: token,
  });
  expect(resCreate.status).toBe(200);
  const taskId = resCreate.body.id;

  // Assign the task
  const resNext = await nextTask({ topic: "InputsTask", token: token });
  expect(resNext.status).toBe(200);

  // Start the task - this creates and runs the flow with the task field data as inputs
  const resStart = await startTask({
    topic: "InputsTask",
    token: token,
    id: taskId,
  });
  expect(resStart.status).toBe(200);
  expect(resStart.body.status).toBe("STARTED");

  const flowRunId = resStart.body.flowRunId;
  expect(flowRunId).toBeDefined();

  // Wait for the flow to complete
  const completedFlow = await flows.inputsTask
    .withAuthToken(token)
    .untilFinished(flowRunId);

  expect(completedFlow.status).toBe("COMPLETED");

  // The flow returns the inputs as data, so we can verify the values were passed correctly
  // The flow implementation also has runtime type assertions that would fail if types are wrong
  // Note: dates come back as Date objects due to the testing-runtime's JSON reviver
  expect(completedFlow.data).toEqual({
    textField: "hello world",
    numberField: 42,
    booleanField: true,
    dateField: new Date("2025-07-15T00:00:00.000Z"),
    timestampField: new Date("2025-07-15T14:30:00.000Z"),
    decimalField: 123.456,
    enumField: "High",
    optionalTextField: "optional value",
  });
});

test("tasks - flow receives null for optional fields", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({
    topic: "InputsTask",
    body: {
      data: {
        textField: "test",
        numberField: 1,
        booleanField: false,
        dateField: "2025-01-01",
        timestampField: "2025-01-01T00:00:00.000Z",
        decimalField: 0.5,
        enumField: "Low",
        // optionalTextField is omitted
      },
    },
    token: token,
  });
  expect(resCreate.status).toBe(200);
  const taskId = resCreate.body.id;

  const resNext = await nextTask({ topic: "InputsTask", token: token });
  expect(resNext.status).toBe(200);

  const resStart = await startTask({
    topic: "InputsTask",
    token: token,
    id: taskId,
  });
  expect(resStart.status).toBe(200);

  const flowRunId = resStart.body.flowRunId;

  const completedFlow = await flows.inputsTask
    .withAuthToken(token)
    .untilFinished(flowRunId);

  expect(completedFlow.status).toBe("COMPLETED");
  expect(completedFlow.data.optionalTextField).toBeNull();
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

async function unassignTask({ topic, token, id }) {
  const url = `${process.env.KEEL_TESTING_API_URL}/topics/json/${topic}/tasks/${id}/unassign`;
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
