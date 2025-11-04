import { resetDatabase, models, flows } from "@teamkeel/testing";
import { isDate } from "node:util/types";
import { beforeEach, expect, test } from "vitest";

beforeEach(resetDatabase);

test("flows - user defined delay retry", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  let flow = await flows.userDelays.withAuthToken(token).start({});

  expect(flow).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "RUNNING",
    startedBy: expect.any(String),
    name: "UserDelays",
    input: {},
    error: null,
    data: null,
    steps: [
      {
        id: expect.any(String),
        name: "user defined delay step",
        runId: expect.any(String),
        stage: null,
        status: "NEW",
        type: "FUNCTION",
        value: null,
        error: null,
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: null,
        endTime: null,
        ui: null,
      },
    ],
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "User delays",
    },
  });

  const updatedFlow = await flows.userDelays
    .withAuthToken(token)
    .untilFinished(flow.id, 20000);

  // We are expecting 3 steps (the initial step + 2 retries)
  expect(updatedFlow).toEqual({
    id: flow.id,
    traceId: flow.traceId,
    status: "COMPLETED",
    name: "UserDelays",
    startedBy: expect.any(String),
    input: {},
    error: null,
    data: null,
    steps: [
      {
        id: flow.steps[0].id,
        name: "user defined delay step",
        runId: flow.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "enforce 3 retries",
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "user defined delay step",
        runId: flow.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "enforce 3 retries",
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "user defined delay step",
        runId: flow.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "enforce 3 retries",
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "user defined delay step",
        runId: flow.id,
        stage: null,
        status: "COMPLETED",
        type: "FUNCTION",
        value: "completed",
        error: null,
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        ui: null,
      },
    ],
    createdAt: flow.createdAt,
    updatedAt: expect.any(Date),
    config: {
      title: "User delays",
    },
  });

  // we have a retry delay policy defined by the user:
  // retry 1, delay 3s
  // retry 2, delay 1s
  // retry 3, delay 2s

  // we have a retryPolicy set as a constant delay of 2s defined on the flow, thus,
  // we expect the second step and third steps to have been delayed by 2 seconds
  // and the third steps to have been delayed by 4 seconds
  let timeDiffMs =
    new Date(updatedFlow.steps[1].startTime as string).getTime() -
    new Date(updatedFlow.steps[1].createdAt as string).getTime();
  expect(timeDiffMs).toBeGreaterThan(3000);
  timeDiffMs =
    new Date(updatedFlow.steps[2].startTime as string).getTime() -
    new Date(updatedFlow.steps[2].createdAt as string).getTime();
  expect(timeDiffMs).toBeGreaterThan(1000);
  timeDiffMs =
    new Date(updatedFlow.steps[3].startTime as string).getTime() -
    new Date(updatedFlow.steps[3].createdAt as string).getTime();
  expect(timeDiffMs).toBeGreaterThan(2000);
});

test("flows - constant delay retry", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  let flow = await flows.delayedRetries.withAuthToken(token).start({});

  expect(flow).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "RUNNING",
    startedBy: expect.any(String),
    name: "DelayedRetries",
    input: {},
    error: null,
    data: null,
    steps: [
      {
        id: expect.any(String),
        name: "constant delay step",
        runId: expect.any(String),
        stage: null,
        status: "NEW",
        type: "FUNCTION",
        value: null,
        error: null,
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: null,
        endTime: null,
        ui: null,
      },
    ],
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "Delayed retries",
    },
  });

  const updatedFlow = await flows.delayedRetries
    .withAuthToken(token)
    .untilFinished(flow.id, 20000);

  // We are expecting 3 steps (the initial step + 2 retries)
  expect(updatedFlow).toEqual({
    id: flow.id,
    traceId: flow.traceId,
    status: "COMPLETED",
    name: "DelayedRetries",
    startedBy: expect.any(String),
    input: {},
    error: null,
    data: null,
    steps: [
      {
        id: flow.steps[0].id,
        name: "constant delay step",
        runId: flow.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "enforce 2 retries",
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "constant delay step",
        runId: flow.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "enforce 2 retries",
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "constant delay step",
        runId: flow.id,
        stage: null,
        status: "COMPLETED",
        type: "FUNCTION",
        value: "completed",
        error: null,
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        ui: null,
      },
    ],
    createdAt: flow.createdAt,
    updatedAt: expect.any(Date),
    config: {
      title: "Delayed retries",
    },
  });

  // we have a retryPolicy set as a constant delay of 2s defined on the flow, thus,
  // we expect the second step and third steps to have been delayed by 2 seconds
  // and the third steps to have been delayed by 4 seconds
  let timeDiffMs =
    new Date(updatedFlow.steps[1].startTime as string).getTime() -
    new Date(updatedFlow.steps[1].createdAt as string).getTime();
  expect(timeDiffMs).toBeGreaterThan(2000);
  timeDiffMs =
    new Date(updatedFlow.steps[2].startTime as string).getTime() -
    new Date(updatedFlow.steps[2].createdAt as string).getTime();
  expect(timeDiffMs).toBeGreaterThan(2000);
});

test("flows - error in step with retries", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  let flow = await flows.errorInStep.withAuthToken(token).start({});

  expect(flow).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "RUNNING",
    startedBy: expect.any(String),
    name: "ErrorInStep",
    input: {},
    error: null,
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
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: null,
        endTime: null,
        ui: null,
      },
    ],
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "Error in step",
    },
  });

  const updatedFlow = await flows.errorInStep
    .withAuthToken(token)
    .untilFinished(flow.id, 20000);

  // We are expecting 3 steps (the initial step + 2 retries)
  expect(updatedFlow).toEqual({
    id: flow.id,
    traceId: flow.traceId,
    status: "FAILED",
    name: "ErrorInStep",
    startedBy: expect.any(String),
    input: {},
    error: "flow failed due to exhausted step retries",
    data: null,
    steps: [
      {
        id: flow.steps[0].id,
        name: "erroring step",
        runId: flow.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "Error in step",
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "erroring step",
        runId: flow.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "Error in step",
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "erroring step",
        runId: flow.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "Error in step",
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "erroring step",
        runId: flow.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "Error in step",
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        ui: null,
      },
    ],
    createdAt: flow.createdAt,
    updatedAt: expect.any(Date),
    config: {
      title: "Error in step",
    },
  });

  // we have a retryPolicy set as a linear backoff with a seed of 2s defined on the flow, thus,
  // we expect the second step to be delayed by 2 seconds
  // and the third steps to have been delayed by 4 seconds
  let timeDiffMs =
    new Date(updatedFlow.steps[1].startTime as string).getTime() -
    new Date(updatedFlow.steps[1].createdAt as string).getTime();
  expect(timeDiffMs).toBeGreaterThan(1000);
  timeDiffMs =
    new Date(updatedFlow.steps[2].startTime as string).getTime() -
    new Date(updatedFlow.steps[2].createdAt as string).getTime();
  expect(timeDiffMs).toBeGreaterThan(2000);
  timeDiffMs =
    new Date(updatedFlow.steps[3].startTime as string).getTime() -
    new Date(updatedFlow.steps[3].createdAt as string).getTime();
  expect(timeDiffMs).toBeGreaterThan(3000);
});

test("flows - on failure callback", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  let flow = await flows.onFailureCallback.withAuthToken(token).start({});

  expect(flow).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "RUNNING",
    startedBy: expect.any(String),
    name: "OnFailureCallback",
    input: {},
    error: null,
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
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: null,
        endTime: null,
        ui: null,
      },
    ],
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "On failure callback",
    },
  });

  const updatedFlow = await flows.onFailureCallback
    .withAuthToken(token)
    .untilFinished(flow.id);

  // We are expecting 2 steps (the initial step + 1 retry)
  expect(updatedFlow).toEqual({
    id: flow.id,
    traceId: flow.traceId,
    status: "FAILED",
    name: "OnFailureCallback",
    startedBy: expect.any(String),
    input: {},
    error: "flow failed due to exhausted step retries",
    data: null,
    steps: [
      {
        id: flow.steps[0].id,
        name: "erroring step",
        runId: flow.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "1 exists",
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "erroring step",
        runId: flow.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "2 exists",
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "erroring step",
        runId: flow.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "3 exists",
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        ui: null,
      },
    ],
    createdAt: flow.createdAt,
    updatedAt: expect.any(Date),
    config: {
      title: "On failure callback",
    },
  });

  expect(await models.thing.findMany()).toHaveLength(0);
});

test("flows - do not retry", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  let flow = await flows.doNotRetry.withAuthToken(token).start({});

  expect(flow).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "RUNNING",
    startedBy: expect.any(String),
    name: "DoNotRetry",
    input: {},
    error: null,
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
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: null,
        endTime: null,
        ui: null,
      },
    ],
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "Do not retry",
    },
  });

  const updatedFlow = await flows.doNotRetry
    .withAuthToken(token)
    .untilFinished(flow.id);

  // We are expecting 2 steps (the initial step + 1 retry)
  expect(updatedFlow).toEqual({
    id: flow.id,
    traceId: flow.traceId,
    status: "FAILED",
    name: "DoNotRetry",
    startedBy: expect.any(String),
    input: {},
    error: "flow failed due to exhausted step retries",
    data: null,
    steps: [
      {
        id: flow.steps[0].id,
        name: "erroring step",
        runId: flow.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "do not retry!",
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        ui: null,
      },
    ],
    createdAt: flow.createdAt,
    updatedAt: expect.any(Date),
    config: {
      title: "Do not retry",
    },
  });

  expect(await models.thing.findMany()).toHaveLength(0);
});

test("flows - eventual step success", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  let flow = await flows.eventualStepSuccess.withAuthToken(token).start({});

  expect(flow).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "RUNNING",
    startedBy: expect.any(String),
    name: "EventualStepSuccess",
    input: {},
    error: null,
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
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: null,
        endTime: null,
        ui: null,
      },
    ],
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "Eventual step success",
    },
  });

  const updatedFlow = await flows.eventualStepSuccess
    .withAuthToken(token)
    .untilFinished(flow.id);

  expect(updatedFlow).toEqual({
    id: flow.id,
    traceId: flow.traceId,
    status: "COMPLETED",
    name: "EventualStepSuccess",
    startedBy: expect.any(String),
    input: {},
    error: null,
    data: null,
    steps: [
      {
        id: flow.steps[0].id,
        name: "erroring step",
        runId: flow.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "Error at attempt 1 of 4",
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "erroring step",
        runId: flow.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "Error at attempt 2 of 4",
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "erroring step",
        runId: flow.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "Error at attempt 3 of 4",
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "erroring step",
        runId: flow.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "Error at attempt 4 of 4",
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "erroring step",
        runId: flow.id,
        stage: null,
        status: "COMPLETED",
        type: "FUNCTION",
        value: "Success at attempt 5",
        error: null,
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        ui: null,
      },
    ],
    createdAt: flow.createdAt,
    updatedAt: expect.any(Date),
    config: {
      title: "Eventual step success",
    },
  });
});

test("flows - error in flow", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  let flow = await flows.errorInFlow.withAuthToken(token).start({});

  expect(flow).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "FAILED",
    name: "ErrorInFlow",
    startedBy: expect.any(String),
    input: {},
    error: "Error in flow",
    data: null,
    steps: [],
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "Error in flow",
    },
  });
});

test("flows - timeout step", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  let flow = await flows.timeoutStep.withAuthToken(token).start({});

  expect(flow).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "RUNNING",
    name: "TimeoutStep",
    startedBy: expect.any(String),
    input: {},
    error: null,
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
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        ui: null,
        startTime: null,
        endTime: null,
      },
    ],
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "Timeout step",
    },
  });

  const updatedFlow = await flows.timeoutStep
    .withAuthToken(token)
    .untilFinished(flow.id);

  // We are expecting 5 steps (the default)
  expect(updatedFlow).toEqual({
    id: flow.id,
    traceId: flow.traceId,
    status: "FAILED",
    name: "TimeoutStep",
    startedBy: expect.any(String),
    input: {},
    error: "flow failed due to exhausted step retries",
    data: null,
    steps: [
      {
        id: flow.steps[0].id,
        name: "timeout step",
        runId: flow.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "Step function timed out after 10ms",
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "timeout step",
        runId: flow.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "Step function timed out after 10ms",
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "timeout step",
        runId: flow.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "Step function timed out after 10ms",
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "timeout step",
        runId: flow.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "Step function timed out after 10ms",
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "timeout step",
        runId: flow.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "Step function timed out after 10ms",
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        ui: null,
      },
    ],
    createdAt: flow.createdAt,
    updatedAt: expect.any(Date),
    config: {
      title: "Timeout step",
    },
  });
});

test("flows - error in validation", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  let flow = await flows.errorInValidation.withAuthToken(token).start({});

  expect(flow.steps[0].status).toBe("PENDING");

  flow = await flows.errorInValidation
    .withAuthToken(token)
    .putStepValues(flow.id, flow.steps[0].id, {});

  expect(flow).toEqual({
    createdAt: expect.any(Date),
    error: "something has gone wrong",
    data: null,
    id: expect.any(String),
    input: {},
    name: "ErrorInValidation",
    startedBy: expect.any(String),
    status: "FAILED",
    steps: [
      {
        createdAt: expect.any(Date),
        endTime: expect.any(Date),
        error: "something has gone wrong",
        id: expect.any(String),
        name: "first page",
        runId: expect.any(String),
        stage: null,
        startTime: expect.any(Date),
        status: "FAILED",
        type: "UI",
        updatedAt: expect.any(Date),
        value: null,
        ui: null,
      },
    ],
    traceId: expect.any(String),
    updatedAt: expect.any(Date),
    config: {
      title: "Error in validation",
    },
  });
});

test("flows - duplicate step name", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  let flow = await flows.duplicateStepName.withAuthToken(token).start({});

  expect(flow).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "RUNNING",
    name: "DuplicateStepName",
    startedBy: expect.any(String),
    input: {},
    error: null,
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
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        endTime: null,
        startTime: null,
      },
    ],
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "Duplicate step name",
    },
  });

  const updatedFlow = await flows.duplicateStepName
    .withAuthToken(token)
    .untilFinished(flow.id);

  expect(updatedFlow).toEqual({
    id: flow.id,
    traceId: flow.traceId,
    status: "FAILED",
    name: "DuplicateStepName",
    startedBy: expect.any(String),
    input: {},
    error: "Duplicate step name: my step",
    data: null,
    steps: [
      {
        id: flow.steps[0].id,
        name: "my step",
        runId: flow.id,
        stage: null,
        status: "COMPLETED",
        type: "FUNCTION",
        value: null,
        error: null,
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "my step",
        runId: flow.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "Duplicate step name: my step",
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        ui: null,
      },
    ],
    createdAt: flow.createdAt,
    updatedAt: expect.any(Date),
    config: {
      title: "Duplicate step name",
    },
  });
});

test("flows - duplicate step name and UI name", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  let flow = await flows.duplicateStepUiName.withAuthToken(token).start({});

  expect(flow).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "RUNNING",
    name: "DuplicateStepUiName",
    startedBy: expect.any(String),
    input: {},
    error: null,
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
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        endTime: null,
        startTime: null,
      },
    ],
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "Duplicate step ui name",
    },
  });

  const updatedFlow = await flows.duplicateStepUiName
    .withAuthToken(token)
    .untilFinished(flow.id);

  expect(updatedFlow).toEqual({
    id: flow.id,
    traceId: flow.traceId,
    status: "FAILED",
    name: "DuplicateStepUiName",
    startedBy: expect.any(String),
    input: {},
    error: "Duplicate step name: my step",
    data: null,
    steps: [
      {
        id: flow.steps[0].id,
        name: "my step",
        runId: flow.id,
        stage: null,
        status: "COMPLETED",
        type: "FUNCTION",
        value: null,
        error: null,
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "my step",
        runId: flow.id,
        stage: null,
        status: "FAILED",
        type: "UI",
        value: null,
        error: "Duplicate step name: my step",
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        ui: null,
      },
    ],
    createdAt: flow.createdAt,
    updatedAt: expect.any(Date),
    config: {
      title: "Duplicate step ui name",
    },
  });
});

test("flows - cancel in-progress steps when flow fails", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  let flow = await flows.cancelInProgressSteps.withAuthToken(token).start({});

  expect(flow).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "RUNNING",
    startedBy: expect.any(String),
    name: "CancelInProgressSteps",
    input: {},
    error: null,
    data: null,
    steps: [
      {
        id: expect.any(String),
        name: "step 1",
        runId: expect.any(String),
        stage: null,
        status: "NEW",
        type: "FUNCTION",
        value: null,
        error: null,
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: null,
        endTime: null,
        ui: null,
      },
    ],
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "Cancel in progress steps",
    },
  });

  const updatedFlow = await flows.cancelInProgressSteps
    .withAuthToken(token)
    .untilFinished(flow.id, 20000);

  // We expect:
  // - step 1: COMPLETED
  // - step 2: FAILED (3 attempts: initial + 2 retries)
  // Note: step 3 is never created because the flow terminates after step 2 exhausts retries
  expect(updatedFlow).toEqual({
    id: flow.id,
    traceId: flow.traceId,
    status: "FAILED",
    name: "CancelInProgressSteps",
    startedBy: expect.any(String),
    input: {},
    error: "flow failed due to exhausted step retries",
    data: null,
    steps: [
      {
        id: flow.steps[0].id,
        name: "step 1",
        runId: flow.id,
        stage: null,
        status: "COMPLETED",
        type: "FUNCTION",
        value: "step1 complete",
        error: null,
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "step 2",
        runId: flow.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "step 2 failed",
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "step 2",
        runId: flow.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "step 2 failed",
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "step 2",
        runId: flow.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "step 2 failed",
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        ui: null,
      },
    ],
    createdAt: flow.createdAt,
    updatedAt: expect.any(Date),
    config: {
      title: "Cancel in progress steps",
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
