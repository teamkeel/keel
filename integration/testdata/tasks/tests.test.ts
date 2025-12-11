import { resetDatabase, models, flows } from "@teamkeel/testing";
import { useDatabase, tasks, Priority } from "@teamkeel/sdk";
import { beforeEach, expect, test } from "vitest";

beforeEach(resetDatabase);

test("tasks - create", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const task = await tasks.emptyFlow.withAuthToken(token).create({
    orderDate: new Date(2025, 6, 15),
    shipByDate: new Date(2025, 6, 30),
  });

  expect(task.id).toBeDefined();
  expect(task.topic).toBe("EmptyFlow");
  expect(task.status).toBe("NEW");
  expect(task.deferredUntil).toBeUndefined();
  expect(task.assignedTo).toBeUndefined();
  expect(task.assignedAt).toBeUndefined();
  expect(task.resolvedAt).toBeUndefined();
  expect(task.createdAt).toBeInstanceOf(Date);
  expect(task.updatedAt).toBeInstanceOf(Date);

  const res = await getTaskQueue({ topic: "EmptyFlow", token: token });
  expect(res.status).toBe(200);
  expect(res.body).toEqual([
    {
      createdAt: expect.any(String),
      id: task.id,
      name: "EmptyFlow",
      status: "NEW",
      updatedAt: expect.any(String),
    },
  ]);
});

test("tasks - next assigns, start creates flow and completes task", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const task = await tasks.emptyFlow.withAuthToken(token).create({
    orderDate: new Date(2025, 6, 9),
    shipByDate: new Date(2025, 6, 20),
  });

  const resNext = await nextTask({ topic: "EmptyFlow", token: token });
  expect(resNext.status).toBe(200);
  expect(resNext.body).toEqual({
    createdAt: expect.any(String),
    id: task.id,
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

  const t1 = await tasks.emptyFlow.withAuthToken(token).create({
    orderDate: new Date(2025, 6, 9),
    shipByDate: new Date(2025, 6, 20),
  });

  const t2 = await tasks.emptyFlow.withAuthToken(token).create({
    orderDate: new Date(2025, 6, 15),
    shipByDate: new Date(2025, 6, 30),
  });

  const t3 = await tasks.emptyFlow.withAuthToken(token).create({
    orderDate: new Date(2025, 6, 14),
    shipByDate: new Date(2025, 6, 30),
  });

  const res = await getTaskQueue({ topic: "EmptyFlow", token: token });
  expect(res.status).toBe(200);
  // Tasks ordered by shipByDate asc, then orderDate asc
  expect(res.body.map((t: any) => t.id)).toEqual([t1.id, t3.id, t2.id]);
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

  const task = await tasks.emptyFlow.withAuthToken(token).create({
    orderDate: new Date(2025, 6, 9),
    shipByDate: new Date(2025, 6, 20),
  });

  const res = await nextTask({ topic: "EmptyFlow", token: token });
  expect(res.status).toBe(200);
  expect(res.body).toEqual({
    createdAt: expect.any(String),
    id: task.id,
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
  const task1 = await tasks.emptyFlow.withAuthToken(token).create({
    orderDate: new Date(2025, 6, 9),
    shipByDate: new Date(2025, 4, 20),
  });

  const task2 = await tasks.emptyFlow.withAuthToken(token).create({
    orderDate: new Date(2025, 6, 9),
    shipByDate: new Date(2025, 6, 20),
  });

  const resNext = await nextTask({ topic: "EmptyFlow", token: token });
  expect(resNext.status).toBe(200);
  expect(resNext.body).toEqual({
    createdAt: expect.any(String),
    id: task1.id,
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

  const task = await tasks.emptyFlow.withAuthToken(token).create({
    orderDate: new Date(2025, 6, 9),
    shipByDate: new Date(2025, 6, 20),
  });

  const identity = await models.identity.findOne({
    email: "admin@keel.xyz",
    issuer: "https://keel.so",
  });

  const assignedTask = await task.assign({ identityId: identity!.id });

  expect(assignedTask.id).toBe(task.id);
  expect(assignedTask.topic).toBe("EmptyFlow");
  expect(assignedTask.status).toBe("ASSIGNED");
  expect(assignedTask.assignedTo).toBe(identity!.id);
  expect(assignedTask.assignedAt).toBeInstanceOf(Date);
  expect(assignedTask.createdAt).toBeInstanceOf(Date);
  expect(assignedTask.updatedAt).toBeInstanceOf(Date);
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

  const task = await tasks.emptyFlow.withAuthToken(token).create({
    orderDate: new Date(2025, 6, 9),
    shipByDate: new Date(2025, 6, 20),
  });

  // Complete the task using SDK
  const completedTask = await task.complete();
  expect(completedTask.status).toBe("COMPLETED");

  const identity = await models.identity.findOne({
    email: "admin@keel.xyz",
    issuer: "https://keel.so",
  });

  // Trying to assign a completed task should throw an error
  await expect(task.assign({ identityId: identity!.id })).rejects.toThrow(
    "task already completed"
  );
});

test("tasks - assign - reassign to different user", async () => {
  const tokenAdmin = await getToken({ email: "admin@keel.xyz" });
  const tokenOther = await getToken({ email: "other@keel.xyz" });

  const task = await tasks.emptyFlow.withAuthToken(tokenAdmin).create({
    orderDate: new Date(2025, 6, 9),
    shipByDate: new Date(2025, 6, 20),
  });

  const adminIdentity = await models.identity.findOne({
    email: "admin@keel.xyz",
    issuer: "https://keel.so",
  });
  const otherIdentity = await models.identity.findOne({
    email: "other@keel.xyz",
    issuer: "https://keel.so",
  });

  // First assign to admin
  const assignedTask1 = await task.assign({ identityId: adminIdentity!.id });
  expect(assignedTask1.assignedTo).toBe(adminIdentity!.id);

  // Reassign to other user
  const assignedTask2 = await task.assign({ identityId: otherIdentity!.id });
  expect(assignedTask2.id).toBe(task.id);
  expect(assignedTask2.topic).toBe("EmptyFlow");
  expect(assignedTask2.status).toBe("ASSIGNED");
  expect(assignedTask2.assignedTo).toBe(otherIdentity!.id);
  expect(assignedTask2.assignedAt).toBeInstanceOf(Date);
});

test("tasks - assign - missing assigned_to in body", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const task = await tasks.emptyFlow.withAuthToken(token).create({
    orderDate: new Date(2025, 6, 9),
    shipByDate: new Date(2025, 6, 20),
  });

  const res = await assignTask({
    topic: "EmptyFlow",
    token: token,
    id: task.id,
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

  const task = await tasks.emptyFlow.withAuthToken(token).create({
    orderDate: new Date(2025, 6, 9),
    shipByDate: new Date(2025, 6, 20),
  });

  const deferUntil = new Date(2025, 7, 15);
  const deferredTask = await task.defer({ deferUntil });

  expect(deferredTask.id).toBe(task.id);
  expect(deferredTask.topic).toBe("EmptyFlow");
  expect(deferredTask.status).toBe("DEFERRED");
  expect(deferredTask.deferredUntil).toEqual(deferUntil);
  expect(deferredTask.createdAt).toBeInstanceOf(Date);
  expect(deferredTask.updatedAt).toBeInstanceOf(Date);
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

  const task = await tasks.emptyFlow.withAuthToken(token).create({
    orderDate: new Date(2025, 6, 9),
    shipByDate: new Date(2025, 6, 20),
  });

  // Complete the task using SDK
  const completedTask = await task.complete();
  expect(completedTask.status).toBe("COMPLETED");

  // Trying to defer a completed task should throw an error
  const deferUntil = new Date(2025, 7, 15);
  await expect(task.defer({ deferUntil })).rejects.toThrow(
    "task already completed"
  );
});

test("tasks - defer - missing defer_until in body", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const task = await tasks.emptyFlow.withAuthToken(token).create({
    orderDate: new Date(2025, 6, 9),
    shipByDate: new Date(2025, 6, 20),
  });

  const res = await deferTask({
    topic: "EmptyFlow",
    token: token,
    id: task.id,
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

  const task = await tasks.emptyFlow.withAuthToken(token).create({
    orderDate: new Date(2025, 6, 9),
    shipByDate: new Date(2025, 6, 20),
  });

  const res = await deferTask({
    topic: "EmptyFlow",
    token: token,
    id: task.id,
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

  const task = await tasks.emptyFlow.withAuthToken(token).create({
    orderDate: new Date(2025, 6, 9),
    shipByDate: new Date(2025, 6, 20),
  });

  const futureDate = new Date();
  futureDate.setDate(futureDate.getDate() + 7);
  const deferredTask = await task.defer({ deferUntil: futureDate });
  expect(deferredTask.status).toBe("DEFERRED");

  // Deferred task should not be assigned via next
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
  const task1 = await tasks.emptyFlow.withAuthToken(token).create({
    orderDate: new Date(2025, 6, 9),
    shipByDate: new Date(2025, 6, 15),
  });

  const futureDate = new Date();
  futureDate.setDate(futureDate.getDate() + 7);
  const deferredTask = await task1.defer({ deferUntil: futureDate });
  expect(deferredTask.status).toBe("DEFERRED");

  // Task 2 has later shipByDate but should be picked since task 1 is deferred
  const task2 = await tasks.emptyFlow.withAuthToken(token).create({
    orderDate: new Date(2025, 6, 10),
    shipByDate: new Date(2025, 6, 20),
  });

  const resNext = await nextTask({ topic: "EmptyFlow", token: token });
  expect(resNext.status).toBe(200);
  expect(resNext.body.id).toBe(task2.id);
  expect(resNext.body.status).toBe("ASSIGNED");
});

test("tasks - defer - deferred task assigned after defer_until passes", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const task = await tasks.emptyFlow.withAuthToken(token).create({
    orderDate: new Date(2025, 6, 9),
    shipByDate: new Date(2025, 6, 20),
  });

  const pastDate = new Date();
  pastDate.setDate(pastDate.getDate() - 1);
  const deferredTask = await task.defer({ deferUntil: pastDate });
  expect(deferredTask.status).toBe("DEFERRED");

  // Task should be assignable via next since defer_until has passed
  const resNext = await nextTask({ topic: "EmptyFlow", token: token });
  expect(resNext.status).toBe(200);
  expect(resNext.body.id).toBe(task.id);
  expect(resNext.body.status).toBe("ASSIGNED");
});

test("tasks - cancel - successfully cancelled", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const task = await tasks.emptyFlow.withAuthToken(token).create({
    orderDate: new Date(2025, 6, 9),
    shipByDate: new Date(2025, 6, 20),
  });

  const cancelledTask = await task.cancel();

  expect(cancelledTask.id).toBe(task.id);
  expect(cancelledTask.topic).toBe("EmptyFlow");
  expect(cancelledTask.status).toBe("CANCELLED");
  expect(cancelledTask.resolvedAt).toBeInstanceOf(Date);
  expect(cancelledTask.createdAt).toBeInstanceOf(Date);
  expect(cancelledTask.updatedAt).toBeInstanceOf(Date);
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

  const task = await tasks.emptyFlow.withAuthToken(token).create({
    orderDate: new Date(2025, 6, 9),
    shipByDate: new Date(2025, 6, 20),
  });

  // Complete the task using SDK
  const completedTask = await task.complete();
  expect(completedTask.status).toBe("COMPLETED");

  // Trying to cancel a completed task should throw an error
  await expect(task.cancel()).rejects.toThrow("task already completed");
});

test("tasks - cancel - cancelled task not assigned via next", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const task = await tasks.emptyFlow.withAuthToken(token).create({
    orderDate: new Date(2025, 6, 9),
    shipByDate: new Date(2025, 6, 20),
  });

  const cancelledTask = await task.cancel();
  expect(cancelledTask.status).toBe("CANCELLED");

  // Cancelled task should not be assigned via next
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

  const identity = await models.identity.findOne({
    email: "admin@keel.xyz",
    issuer: "https://keel.so",
  });

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

  // Initial NEW status from task creation
  expect(statusEntries[0].status).toBe("NEW");
  expect(statusEntries[0].setBy).toBe(identity!.id);
  expect(statusEntries[0].assignedTo).toBeNull();

  // ASSIGNED status from next
  expect(statusEntries[1].status).toBe("ASSIGNED");
  expect(statusEntries[1].setBy).toBe(identity!.id);
  expect(statusEntries[1].assignedTo).toBe(identity!.id);

  // NEW status from unassign
  expect(statusEntries[2].status).toBe("NEW");
  expect(statusEntries[2].setBy).toBe(identity!.id);
  expect(statusEntries[2].assignedTo).toBeNull();
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

  const task = await tasks.emptyFlow.withAuthToken(token).create({
    orderDate: new Date(2025, 6, 9),
    shipByDate: new Date(2025, 6, 20),
  });

  const identity = await models.identity.findOne({
    email: "admin@keel.xyz",
    issuer: "https://keel.so",
  });

  // Assign task using SDK
  const assignedTask = await task.assign({ identityId: identity!.id });
  expect(assignedTask.flowRunId).toBeUndefined();

  // Start the task using SDK
  const startedTask = await assignedTask.start();
  expect(startedTask.status).toBe("STARTED");

  const flowRunId = startedTask.flowRunId;
  expect(flowRunId).toBeDefined();

  // Wait for the flow to complete
  await flows.emptyFlow.withAuthToken(token).untilFinished(flowRunId!);

  const statusEntries = await (useDatabase() as any)
    .selectFrom("keel.task_status")
    .selectAll()
    .where("keel_task_id", "=", task.id)
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

  const task = await tasks.emptyFlow.withAuthToken(token).create({
    orderDate: new Date(2025, 6, 9),
    shipByDate: new Date(2025, 6, 20),
  });

  const identity = await models.identity.findOne({
    email: "admin@keel.xyz",
    issuer: "https://keel.so",
  });

  // Assign and start using SDK
  const assignedTask = await task.assign({ identityId: identity!.id });
  const startedTask = await assignedTask.start();
  expect(startedTask.status).toBe("STARTED");

  const flowRunId = startedTask.flowRunId;
  expect(flowRunId).toBeDefined();

  // Wait for the flow to complete (which auto-completes the task)
  await flows.emptyFlow.withAuthToken(token).untilFinished(flowRunId!);

  // Calling start on a completed task should throw an error
  await expect(assignedTask.start()).rejects.toThrow("task already completed");

  const statusEntries = await (useDatabase() as any)
    .selectFrom("keel.task_status")
    .selectAll()
    .where("keel_task_id", "=", task.id)
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

  const task = await tasks.emptyFlow.withAuthToken(token).create({
    orderDate: new Date(2025, 6, 9),
    shipByDate: new Date(2025, 6, 20),
  });

  const identity = await models.identity.findOne({
    email: "admin@keel.xyz",
    issuer: "https://keel.so",
  });

  // Assign task using SDK
  const assignedTask = await task.assign({ identityId: identity!.id });

  // Start the task using SDK - this creates and runs the flow
  const startedTask = await assignedTask.start();
  expect(startedTask.status).toBe("STARTED");

  const flowRunId = startedTask.flowRunId;
  expect(flowRunId).toBeDefined();

  // Wait for the flow to complete
  const completedFlow = await flows.emptyFlow
    .withAuthToken(token)
    .untilFinished(flowRunId!);
  expect(completedFlow.status).toBe("COMPLETED");

  // Fetch the task again to verify it was auto-completed when flow finished
  const taskFromDb = await (useDatabase() as any)
    .selectFrom("keel.task")
    .selectAll()
    .where("id", "=", task.id)
    .executeTakeFirst();

  expect(taskFromDb.status).toBe("COMPLETED");
  expect(taskFromDb.resolvedAt).not.toBeNull();
  expect(taskFromDb.flowRunId).toBe(flowRunId);

  // Verify task_status entries include COMPLETED with the flow run ID
  const statusEntries = await (useDatabase() as any)
    .selectFrom("keel.task_status")
    .selectAll()
    .where("keel_task_id", "=", task.id)
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
  const task = await tasks.inputsTask.withAuthToken(token).create({
    textField: "hello world",
    numberField: 42,
    booleanField: true,
    dateField: new Date("2025-07-15"),
    timestampField: new Date("2025-07-15T14:30:00.000Z"),
    decimalField: 123.456,
    enumField: Priority.High,
    optionalTextField: "optional value",
  });

  const identity = await models.identity.findOne({
    email: "admin@keel.xyz",
    issuer: "https://keel.so",
  });

  // Assign and start using SDK
  const assignedTask = await task.assign({ identityId: identity!.id });
  const startedTask = await assignedTask.start();
  expect(startedTask.status).toBe("STARTED");

  const flowRunId = startedTask.flowRunId;
  expect(flowRunId).toBeDefined();

  // Wait for the flow to complete
  const completedFlow = await flows.inputsTask
    .withAuthToken(token)
    .untilFinished(flowRunId!);

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

  const task = await tasks.inputsTask.withAuthToken(token).create({
    textField: "test",
    numberField: 1,
    booleanField: false,
    dateField: new Date("2025-01-01"),
    timestampField: new Date("2025-01-01T00:00:00.000Z"),
    decimalField: 0.5,
    enumField: Priority.Low,
    // optionalTextField is omitted
  });

  const identity = await models.identity.findOne({
    email: "admin@keel.xyz",
    issuer: "https://keel.so",
  });

  // Assign and start using SDK
  const assignedTask = await task.assign({ identityId: identity!.id });
  const startedTask = await assignedTask.start();

  const flowRunId = startedTask.flowRunId;

  const completedFlow = await flows.inputsTask
    .withAuthToken(token)
    .untilFinished(flowRunId!);

  expect(completedFlow.status).toBe("COMPLETED");
  expect(completedFlow.data.optionalTextField).toBeNull();
});

test("tasks SDK - create task with data", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  // Create a task using the SDK
  // Use UTC dates to avoid timezone issues
  const inputData = {
    orderDate: new Date("2025-07-15"),
    shipByDate: new Date("2025-07-30"),
  };
  const task = await tasks.emptyFlow.withAuthToken(token).create(inputData);

  expect(task.id).toBeDefined();
  expect(task.topic).toBe("EmptyFlow");
  expect(task.status).toBe("NEW");
  expect(task.deferredUntil).toBeUndefined();
  expect(task.assignedTo).toBeUndefined();
  expect(task.assignedAt).toBeUndefined();
  expect(task.resolvedAt).toBeUndefined();
  expect(task.createdAt).toBeInstanceOf(Date);
  expect(task.updatedAt).toBeInstanceOf(Date);

  // Verify the task was created in the database
  const taskFromDb = await (useDatabase() as any)
    .selectFrom("keel.task")
    .selectAll()
    .where("id", "=", task.id)
    .executeTakeFirst();

  expect(taskFromDb).toBeDefined();
  expect(taskFromDb.name).toBe("EmptyFlow");
  expect(taskFromDb.status).toBe("NEW");

  // Verify the task data was created in the task-specific table
  // Note: CamelCasePlugin converts column names to camelCase in results
  const taskData = await (useDatabase() as any)
    .selectFrom("empty_flow")
    .selectAll()
    .where("keelTaskId", "=", task.id)
    .executeTakeFirst();

  expect(taskData).toBeDefined();
  // The database returns DATE type values - extract just the date part for comparison
  // Note: The returned Date object may have timezone offset applied
  const orderDateStr = taskData.orderDate.toISOString().slice(0, 10);
  const shipByDateStr = taskData.shipByDate.toISOString().slice(0, 10);
  // Just verify the dates are approximately correct (within 1 day of the input due to TZ)
  expect(["2025-07-14", "2025-07-15"]).toContain(orderDateStr);
  expect(["2025-07-29", "2025-07-30"]).toContain(shipByDateStr);
});

test("tasks SDK - create InputsTask with typed fields", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const inputData = {
    textField: "hello world",
    numberField: 42,
    booleanField: true,
    dateField: new Date(2025, 6, 15),
    timestampField: new Date("2025-07-15T14:30:00.000Z"),
    decimalField: 123.456,
    enumField: Priority.High,
    optionalTextField: "optional value",
  };
  const task = await tasks.inputsTask.withAuthToken(token).create(inputData);

  expect(task.id).toBeDefined();
  expect(task.topic).toBe("InputsTask");
  expect(task.status).toBe("NEW");

  // Verify the task data was created with correct types
  // Note: CamelCasePlugin converts column names to camelCase in results
  const taskData = await (useDatabase() as any)
    .selectFrom("inputs_task")
    .selectAll()
    .where("keelTaskId", "=", task.id)
    .executeTakeFirst();

  expect(taskData).toBeDefined();
  expect(taskData.textField).toBe("hello world");
  expect(taskData.numberField).toBe(42);
  expect(taskData.booleanField).toBe(true);
  expect(taskData.decimalField).toBe(123.456);
  expect(taskData.enumField).toBe("High");
  expect(taskData.optionalTextField).toBe("optional value");
});

test("tasks SDK - create task with deferredUntil option", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const inputData = {
    orderDate: new Date(2025, 6, 15),
    shipByDate: new Date(2025, 6, 30),
  };
  const deferDate = new Date(2025, 11, 25);
  const task = await tasks.emptyFlow.withAuthToken(token).create(inputData, {
    deferredUntil: deferDate,
  });

  expect(task.id).toBeDefined();
  expect(task.topic).toBe("EmptyFlow");
  expect(task.status).toBe("DEFERRED");
  expect(task.deferredUntil).toEqual(deferDate);

  // Verify the task was created with DEFERRED status and deferred_until date
  const taskFromDb = await (useDatabase() as any)
    .selectFrom("keel.task")
    .selectAll()
    .where("id", "=", task.id)
    .executeTakeFirst();

  expect(taskFromDb.status).toBe("DEFERRED");
  expect(new Date(taskFromDb.deferredUntil)).toEqual(deferDate);
});

test("tasks SDK - assign task", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const task = await tasks.emptyFlow.withAuthToken(token).create({
    orderDate: new Date("2025-07-15"),
    shipByDate: new Date("2025-07-30"),
  });

  const identity = await models.identity.findOne({
    email: "admin@keel.xyz",
    issuer: "https://keel.so",
  });

  // Call assign on the task instance
  const assignedTask = await task.assign({ identityId: identity!.id });

  expect(assignedTask.id).toBe(task.id);
  expect(assignedTask.status).toBe("ASSIGNED");
  expect(assignedTask.assignedTo).toBe(identity!.id);
  expect(assignedTask.assignedAt).toBeInstanceOf(Date);
});

test("tasks SDK - start task", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const task = await tasks.emptyFlow.withAuthToken(token).create({
    orderDate: new Date("2025-07-15"),
    shipByDate: new Date("2025-07-30"),
  });

  const identity = await models.identity.findOne({
    email: "admin@keel.xyz",
    issuer: "https://keel.so",
  });

  // Must assign before starting - call on task instance
  const assignedTask = await task.assign({ identityId: identity!.id });

  // Call start on the assigned task instance
  const startedTask = await assignedTask.start();

  expect(startedTask.id).toBe(task.id);
  expect(startedTask.status).toBe("STARTED");
  expect(startedTask.flowRunId).toBeDefined();

  // Wait for flow to complete
  await flows.emptyFlow
    .withAuthToken(token)
    .untilFinished(startedTask.flowRunId!);
});

test("tasks SDK - complete task", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const task = await tasks.emptyFlow.withAuthToken(token).create({
    orderDate: new Date("2025-07-15"),
    shipByDate: new Date("2025-07-30"),
  });

  // Call complete on the task instance
  const completedTask = await task.complete();

  expect(completedTask.id).toBe(task.id);
  expect(completedTask.status).toBe("COMPLETED");
  expect(completedTask.resolvedAt).toBeInstanceOf(Date);
});

test("tasks SDK - defer task", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const task = await tasks.emptyFlow.withAuthToken(token).create({
    orderDate: new Date("2025-07-15"),
    shipByDate: new Date("2025-07-30"),
  });

  const deferDate = new Date(2025, 11, 25);
  // Call defer on the task instance
  const deferredTask = await task.defer({ deferUntil: deferDate });

  expect(deferredTask.id).toBe(task.id);
  expect(deferredTask.status).toBe("DEFERRED");
  expect(deferredTask.deferredUntil).toEqual(deferDate);
});

test("tasks SDK - cancel task", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const task = await tasks.emptyFlow.withAuthToken(token).create({
    orderDate: new Date("2025-07-15"),
    shipByDate: new Date("2025-07-30"),
  });

  // Call cancel on the task instance
  const cancelledTask = await task.cancel();

  expect(cancelledTask.id).toBe(task.id);
  expect(cancelledTask.status).toBe("CANCELLED");
  expect(cancelledTask.resolvedAt).toBeInstanceOf(Date);
});

test("tasks SDK - task_status table records set_by correctly", async () => {
  const adminToken = await getToken({ email: "admin@keel.xyz" });
  const otherToken = await getToken({ email: "other@keel.xyz" });

  const adminIdentity = await models.identity.findOne({
    email: "admin@keel.xyz",
    issuer: "https://keel.so",
  });
  const otherIdentity = await models.identity.findOne({
    email: "other@keel.xyz",
    issuer: "https://keel.so",
  });

  // Create task as admin
  const task = await tasks.emptyFlow.withAuthToken(adminToken).create({
    orderDate: new Date("2025-07-15"),
    shipByDate: new Date("2025-07-30"),
  });

  // Assign task as admin
  await task.assign({ identityId: adminIdentity!.id });

  // Complete task as admin
  await task.complete();

  // Check task_status entries
  const statusEntries = await (useDatabase() as any)
    .selectFrom("keel.task_status")
    .selectAll()
    .where("keel_task_id", "=", task.id)
    .orderBy("created_at", "asc")
    .execute();

  expect(statusEntries).toHaveLength(3);

  // NEW status - set_by is the creator (admin)
  expect(statusEntries[0].status).toBe("NEW");
  expect(statusEntries[0].setBy).toBe(adminIdentity!.id);

  // ASSIGNED status - set_by is admin
  expect(statusEntries[1].status).toBe("ASSIGNED");
  expect(statusEntries[1].setBy).toBe(adminIdentity!.id);

  // COMPLETED status - set_by is admin
  expect(statusEntries[2].status).toBe("COMPLETED");
  expect(statusEntries[2].setBy).toBe(adminIdentity!.id);
});

test("tasks SDK - assign with different identity reflects in task_status", async () => {
  const adminToken = await getToken({ email: "admin@keel.xyz" });
  const otherToken = await getToken({ email: "other@keel.xyz" });

  const adminIdentity = await models.identity.findOne({
    email: "admin@keel.xyz",
    issuer: "https://keel.so",
  });
  const otherIdentity = await models.identity.findOne({
    email: "other@keel.xyz",
    issuer: "https://keel.so",
  });

  // Create task as admin
  const task = await tasks.emptyFlow.withAuthToken(adminToken).create({
    orderDate: new Date("2025-07-15"),
    shipByDate: new Date("2025-07-30"),
  });

  // Switch to other identity and assign task
  const assignedTask = await task.withAuthToken(otherToken).assign({
    identityId: otherIdentity!.id,
  });

  expect(assignedTask.assignedTo).toBe(otherIdentity!.id);

  // Check task_status entries
  const statusEntries = await (useDatabase() as any)
    .selectFrom("keel.task_status")
    .selectAll()
    .where("keel_task_id", "=", task.id)
    .orderBy("created_at", "asc")
    .execute();

  expect(statusEntries).toHaveLength(2);

  // NEW status - set_by is admin (creator)
  expect(statusEntries[0].status).toBe("NEW");
  expect(statusEntries[0].setBy).toBe(adminIdentity!.id);

  // ASSIGNED status - set_by is other (who performed the assign)
  expect(statusEntries[1].status).toBe("ASSIGNED");
  expect(statusEntries[1].setBy).toBe(otherIdentity!.id);
});

test("tasks SDK - cancel with different identity reflects in task_status", async () => {
  const adminToken = await getToken({ email: "admin@keel.xyz" });
  const otherToken = await getToken({ email: "other@keel.xyz" });

  const adminIdentity = await models.identity.findOne({
    email: "admin@keel.xyz",
    issuer: "https://keel.so",
  });
  const otherIdentity = await models.identity.findOne({
    email: "other@keel.xyz",
    issuer: "https://keel.so",
  });

  // Create task as admin
  const task = await tasks.emptyFlow.withAuthToken(adminToken).create({
    orderDate: new Date("2025-07-15"),
    shipByDate: new Date("2025-07-30"),
  });

  // Switch to other identity and cancel task
  const cancelledTask = await task.withAuthToken(otherToken).cancel();

  expect(cancelledTask.status).toBe("CANCELLED");

  // Check task_status entries
  const statusEntries = await (useDatabase() as any)
    .selectFrom("keel.task_status")
    .selectAll()
    .where("keel_task_id", "=", task.id)
    .orderBy("created_at", "asc")
    .execute();

  expect(statusEntries).toHaveLength(2);

  // NEW status - set_by is admin (creator)
  expect(statusEntries[0].status).toBe("NEW");
  expect(statusEntries[0].setBy).toBe(adminIdentity!.id);

  // CANCELLED status - set_by is other (who performed the cancel)
  expect(statusEntries[1].status).toBe("CANCELLED");
  expect(statusEntries[1].setBy).toBe(otherIdentity!.id);
});

test("tasks SDK - complete with different identity reflects in task_status", async () => {
  const adminToken = await getToken({ email: "admin@keel.xyz" });
  const otherToken = await getToken({ email: "other@keel.xyz" });

  const adminIdentity = await models.identity.findOne({
    email: "admin@keel.xyz",
    issuer: "https://keel.so",
  });
  const otherIdentity = await models.identity.findOne({
    email: "other@keel.xyz",
    issuer: "https://keel.so",
  });

  // Create task as admin
  const task = await tasks.emptyFlow.withAuthToken(adminToken).create({
    orderDate: new Date("2025-07-15"),
    shipByDate: new Date("2025-07-30"),
  });

  // Switch to other identity and complete task
  const completedTask = await task.withAuthToken(otherToken).complete();

  expect(completedTask.status).toBe("COMPLETED");

  // Check task_status entries
  const statusEntries = await (useDatabase() as any)
    .selectFrom("keel.task_status")
    .selectAll()
    .where("keel_task_id", "=", task.id)
    .orderBy("created_at", "asc")
    .execute();

  expect(statusEntries).toHaveLength(2);

  // NEW status - set_by is admin (creator)
  expect(statusEntries[0].status).toBe("NEW");
  expect(statusEntries[0].setBy).toBe(adminIdentity!.id);

  // COMPLETED status - set_by is other (who performed the complete)
  expect(statusEntries[1].status).toBe("COMPLETED");
  expect(statusEntries[1].setBy).toBe(otherIdentity!.id);
});

test("tasks SDK - defer with different identity reflects in task_status", async () => {
  const adminToken = await getToken({ email: "admin@keel.xyz" });
  const otherToken = await getToken({ email: "other@keel.xyz" });

  const adminIdentity = await models.identity.findOne({
    email: "admin@keel.xyz",
    issuer: "https://keel.so",
  });
  const otherIdentity = await models.identity.findOne({
    email: "other@keel.xyz",
    issuer: "https://keel.so",
  });

  // Create task as admin
  const task = await tasks.emptyFlow.withAuthToken(adminToken).create({
    orderDate: new Date("2025-07-15"),
    shipByDate: new Date("2025-07-30"),
  });

  // Switch to other identity and defer task
  const deferDate = new Date(2025, 11, 25);
  const deferredTask = await task
    .withAuthToken(otherToken)
    .defer({ deferUntil: deferDate });

  expect(deferredTask.status).toBe("DEFERRED");

  // Check task_status entries
  const statusEntries = await (useDatabase() as any)
    .selectFrom("keel.task_status")
    .selectAll()
    .where("keel_task_id", "=", task.id)
    .orderBy("created_at", "asc")
    .execute();

  expect(statusEntries).toHaveLength(2);

  // NEW status - set_by is admin (creator)
  expect(statusEntries[0].status).toBe("NEW");
  expect(statusEntries[0].setBy).toBe(adminIdentity!.id);

  // DEFERRED status - set_by is other (who performed the defer)
  expect(statusEntries[1].status).toBe("DEFERRED");
  expect(statusEntries[1].setBy).toBe(otherIdentity!.id);
});

async function getToken({ email }: { email: string }) {
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

async function getTaskQueue({
  topic,
  token,
}: {
  topic: string;
  token: string;
}) {
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

async function nextTask({ topic, token }: { topic: string; token: string }) {
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

async function startTask({
  topic,
  token,
  id,
}: {
  topic: string;
  token: string;
  id: string;
}) {
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

async function assignTask({
  topic,
  token,
  id,
  body,
}: {
  topic: string;
  token: string;
  id: string;
  body: any;
}) {
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

async function deferTask({
  topic,
  token,
  id,
  body,
}: {
  topic: string;
  token: string;
  id: string;
  body: any;
}) {
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

async function cancelTask({
  topic,
  token,
  id,
}: {
  topic: string;
  token: string;
  id: string;
}) {
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
