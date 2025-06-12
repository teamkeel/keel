import { resetDatabase, models } from "@teamkeel/testing";
import { MyEnum } from "@teamkeel/sdk";
import { beforeEach, expect, test } from "vitest";

beforeEach(resetDatabase);

/*
TEST CASES
========================================
[x] Stepless flow function
[x] Flow function with consecutive UI steps
[x] Flow function with consecutive function steps
[x] Flow function with alternating function and UI steps
[x] Error thrown in flow function
[x] Error thrown in step function
[x] Step returning scalar value
[x] Step function retrying  
[x] Step function timing out 
[x] UI step validation
[x] Check full API responses
[ ] All UI elements response
[x] Test all Keel types as inputs
[x] Permissions and identity tests
[ ] Stages
[x] List my runs
[x] ctx env
*/

test("flows - scalar step", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  let { status, body } = await startFlow({
    name: "scalarStep",
    token,
    body: {},
  });
  expect(status).toEqual(200);

  const flow = await untilFlowFinished({
    name: "scalarStep",
    id: body.id,
    token,
  });

  expect(flow).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "COMPLETED",
    name: "ScalarStep",
    input: {},
    startedBy: expect.any(String),
    steps: [
      {
        id: expect.any(String),
        name: "scalar step",
        runId: expect.any(String),
        stage: null,
        status: "COMPLETED",
        type: "FUNCTION",
        value: 10,
        error: null,
        startTime: expect.any(String),
        endTime: expect.any(String),
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        ui: null,
      },
    ],
    createdAt: expect.any(String),
    updatedAt: expect.any(String),
    config: null,
  });
});

test("flows - only functions with config", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  let { status, body } = await startFlow({
    name: "onlyFunctions",
    token,
    body: {
      name: "My Thing",
      age: 25,
    },
  });
  expect(status).toEqual(200);

  expect(body).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "RUNNING",
    name: "OnlyFunctions",
    startedBy: expect.any(String),
    input: {
      name: "My Thing",
      age: 25,
    },
    steps: [
      {
        id: expect.any(String),
        name: "insert thing",
        runId: expect.any(String),
        stage: "stage1",
        status: "NEW",
        type: "FUNCTION",
        value: null,
        error: null,
        startTime: null,
        endTime: null,
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        ui: null,
      },
    ],
    createdAt: expect.any(String),
    updatedAt: expect.any(String),
    config: {
      description: "This is a description",
      stages: [
        {
          description: "This is stage 1's description",
          key: "stage1",
          name: "My stage 1",
        },
        {
          description: "This is stage 2's description",
          key: "stage2",
          name: "My stage 2",
        },
      ],
      title: "Flow with two functions",
    },
  });

  const flow = await untilFlowFinished({
    name: "onlyFunctions",
    id: body.id,
    token,
  });

  expect(flow).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "COMPLETED",
    name: "OnlyFunctions",
    startedBy: expect.any(String),
    input: {
      name: "My Thing",
      age: 25,
    },
    steps: [
      {
        id: expect.any(String),
        name: "insert thing",
        runId: expect.any(String),
        stage: "stage1",
        status: "COMPLETED",
        type: "FUNCTION",
        value: expect.any(String),
        error: null,
        startTime: expect.any(String),
        endTime: expect.any(String),
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "update thing",
        runId: expect.any(String),
        stage: "stage2",
        status: "COMPLETED",
        type: "FUNCTION",
        value: {
          name: "My Thing Updated",
          age: 26,
        },
        error: null,
        startTime: expect.any(String),
        endTime: expect.any(String),
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        ui: null,
      },
    ],
    createdAt: expect.any(String),
    updatedAt: expect.any(String),
    config: null,
  });
});

test("flows - only pages", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  let { status, body } = await startFlow({
    name: "onlyPages",
    token,
    body: {},
  });
  expect(status).toEqual(200);
  expect(body).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "AWAITING_INPUT",
    name: "OnlyPages",
    startedBy: expect.any(String),
    input: {},
    steps: [
      {
        id: expect.any(String),
        name: "first page",
        runId: expect.any(String),
        stage: null,
        status: "PENDING",
        type: "UI",
        value: null,
        error: null,
        startTime: expect.any(String),
        endTime: null,
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        ui: {
          __type: "ui.page",
          content: [
            { __type: "ui.display.grid", data: [{ title: "A thing" }] },
          ],
          title: "Grid of things",
        },
      },
    ],
    createdAt: expect.any(String),
    updatedAt: expect.any(String),
    config: {
      title: "Only pages",
    },
  });

  // Provide the values for the pending UI step
  ({ status, body } = await putStepValues({
    name: "MixedStepTypes",
    runId: body.id,
    stepId: body.steps[0].id,
    token,
    values: {},
  }));
  expect(status).toEqual(200);
  expect(body).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "AWAITING_INPUT",
    name: "OnlyPages",
    startedBy: expect.any(String),
    input: {},
    steps: [
      {
        id: expect.any(String),
        name: "first page",
        runId: expect.any(String),
        stage: null,
        status: "COMPLETED",
        type: "UI",
        value: {},
        error: null,
        startTime: expect.any(String),
        endTime: expect.any(String),
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "question",
        runId: expect.any(String),
        stage: null,
        status: "PENDING",
        type: "UI",
        value: null,
        error: null,
        startTime: expect.any(String),
        endTime: null,
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        ui: {
          __type: "ui.page",
          content: [
            {
              __type: "ui.input.boolean",
              label: "Did you like the things?",
              disabled: false,
              mode: "checkbox",
              name: "yesno",
              optional: false,
            },
          ],
          title: "My flow",
        },
      },
    ],
    createdAt: expect.any(String),
    updatedAt: expect.any(String),
    config: {
      title: "Only pages",
    },
  });

  ({ status, body } = await putStepValues({
    name: "MixedStepTypes",
    runId: body.id,
    stepId: body.steps[1].id,
    token,
    values: { yesno: true },
  }));
  expect(status).toEqual(200);
  expect(body).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "COMPLETED",
    name: "OnlyPages",
    startedBy: expect.any(String),
    input: {},
    steps: [
      {
        id: expect.any(String),
        name: "first page",
        runId: expect.any(String),
        stage: null,
        status: "COMPLETED",
        type: "UI",
        value: {},
        error: null,
        startTime: expect.any(String),
        endTime: expect.any(String),
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "question",
        runId: expect.any(String),
        stage: null,
        status: "COMPLETED",
        type: "UI",
        value: { yesno: true },
        error: null,
        startTime: expect.any(String),
        endTime: expect.any(String),
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        ui: null,
      },
    ],
    createdAt: expect.any(String),
    updatedAt: expect.any(String),
    config: {
      title: "Only pages",
    },
  });
});

test("flows - stepless flow", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  let { status, body } = await startFlow({
    name: "Stepless",
    token,
    body: {},
  });
  expect(status).toEqual(200);

  const flow = await untilFlowFinished({
    name: "Stepless",
    id: body.id,
    token,
  });

  // Flow has no steps so should be synchronously completed
  expect(flow).toEqual({
    id: body.id,
    input: {},
    name: "Stepless",
    startedBy: expect.any(String),
    status: "COMPLETED",
    steps: [],
    config: null,
    traceId: expect.any(String),
    createdAt: expect.any(String),
    updatedAt: expect.any(String),
  });

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

  // First step is a function so should be in status NEW - it will get run async via the queue
  expect(body).toEqual({
    id: expect.any(String),
    input: {},
    name: "SingleStep",
    startedBy: expect.any(String),
    status: "RUNNING",
    steps: [
      {
        id: expect.any(String),
        runId: body.id,
        stage: null,
        name: "insert thing",
        error: null,
        ui: null,
        status: "NEW",
        type: "FUNCTION",
        value: null,
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        startTime: null,
        endTime: null,
      },
    ],
    config: {
      title: "Single step",
    },
    traceId: expect.any(String),
    createdAt: expect.any(String),
    updatedAt: expect.any(String),
  });

  const flow = await untilFlowFinished({
    name: "singleStep",
    id: body.id,
    token,
  });

  // Now the flow has finished the run and step statuses should have been
  // updated and the returned value stored against the step
  expect(flow).toEqual({
    id: body.id,
    input: {},
    name: "SingleStep",
    startedBy: expect.any(String),
    status: "COMPLETED",
    steps: [
      {
        id: body.steps[0].id,
        runId: body.id,
        stage: null,
        name: "insert thing",
        error: null,
        ui: null,
        status: "COMPLETED",
        type: "FUNCTION",
        value: {
          number: 10,
        },
        createdAt: body.steps[0].createdAt,
        updatedAt: expect.any(String),
        startTime: expect.any(String),
        endTime: expect.any(String),
      },
    ],
    config: null,
    traceId: body.traceId,
    createdAt: body.createdAt,
    updatedAt: expect.any(String),
  });
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

  // First step is a function so API response should show that as NEW
  expect(body).toEqual({
    id: expect.any(String),
    input: {
      name: "Keelson",
      age: 23,
    },
    name: "MixedStepTypes",
    startedBy: expect.any(String),
    status: "RUNNING",
    steps: [
      {
        id: expect.any(String),
        runId: body.id,
        stage: null,
        name: "insert thing",
        error: null,
        ui: null,
        status: "NEW",
        type: "FUNCTION",
        value: null,
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        startTime: null,
        endTime: null,
      },
    ],
    config: {
      title: "Mixed step types",
    },
    traceId: expect.any(String),
    createdAt: expect.any(String),
    updatedAt: expect.any(String),
  });

  const runId = body.id;
  const traceId = body.traceId;
  let step1 = body.steps[0];

  // The second step is a page with UI so we wait until the flow has reached that point
  body = await untilFlowAwaitingInput({
    name: "MixedStepTypes",
    id: runId,
    token,
  });
  expect(body).toEqual({
    id: runId,
    name: "MixedStepTypes",
    startedBy: expect.any(String),
    status: "AWAITING_INPUT", // Flow is now awaiting input
    input: {
      name: "Keelson",
      age: 23,
    },
    config: {
      title: "Mixed step types",
    },
    traceId,
    createdAt: expect.any(String),
    updatedAt: expect.any(String),
    steps: [
      {
        id: step1.id,
        runId: runId,
        stage: null,
        name: "insert thing",
        error: null,
        ui: null,
        status: "COMPLETED", // First step has been completed
        type: "FUNCTION",
        // We have the value stored from this step now
        value: {
          id: expect.any(String),
        },
        createdAt: step1.createdAt,
        updatedAt: expect.any(String),
        startTime: expect.any(String),
        endTime: expect.any(String),
      },
      {
        id: expect.any(String),
        runId: runId,
        stage: null,
        name: "confirm thing",
        error: null,
        // We have the full UI config because this step is awaiting user input
        ui: {
          __type: "ui.page",
          title: "Update thing",
          description: "Confirm the existing data in thing",
          content: [
            {
              __type: "ui.input.text",
              defaultValue: "Keelson",
              disabled: false,
              label: "Name",
              name: "name",
              optional: false,
            },
            {
              __type: "ui.display.divider",
            },
            {
              __type: "ui.input.number",
              defaultValue: 23,
              disabled: false,
              label: "Age",
              name: "age",
              optional: false,
            },
          ],
        },
        status: "PENDING", // This step is now pending while it waits for user input
        type: "UI",
        value: null,
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        startTime: expect.any(String),
        endTime: null,
      },
    ],
  });

  step1 = body.steps[0];
  let step2 = body.steps[1];

  // The first step created a db record so we check that exists
  let thing = await models.thing.findOne({
    id: step1.value.id,
  });
  expect(thing!.name).toBe("Keelson");
  expect(thing!.age).toBe(23);

  // Provide the values for the pending UI step
  ({ status, body } = await putStepValues({
    name: "MixedStepTypes",
    runId,
    stepId: step2.id,
    token,
    values: {
      name: "Keelson updated",
      age: 32,
    },
  }));
  expect(status).toEqual(200);
  expect(body).toEqual({
    id: runId,
    name: "MixedStepTypes",
    startedBy: expect.any(String),
    status: "RUNNING",
    input: {
      name: "Keelson",
      age: 23,
    },
    steps: [
      step1, // Step 1 should not have changed
      {
        // Now this step has been completed because we provided the values, the ui
        // config is no longer present, the status is COMPLETED and we have the stored values
        ...step2,
        ui: null,
        status: "COMPLETED",
        value: {
          name: "Keelson updated",
          age: 32,
        },
        updatedAt: expect.any(String),
        endTime: expect.any(String),
      },
      {
        // The final step is now pending and will be run via the queue
        id: expect.any(String),
        runId,
        stage: null,
        status: "NEW",
        type: "FUNCTION",
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        name: "update thing",
        error: null,
        ui: null,
        value: null,
        startTime: null,
        endTime: null,
      },
    ],
    config: {
      title: "Mixed step types",
    },
    traceId,
    createdAt: expect.any(String),
    updatedAt: expect.any(String),
  });

  step2 = body.steps[1];
  let step3 = body.steps[2];

  body = await untilFlowFinished({ name: "mixedStepTypes", id: runId, token });
  expect(body.status).toBe("COMPLETED");
  expect(body.steps[2]).toEqual({
    // The final step is now complete and will contain the result
    id: step3.id,
    runId,
    status: "COMPLETED",
    type: "FUNCTION",
    name: "update thing",
    error: null,
    ui: null,
    value: {
      name: "Keelson updated",
      age: 32,
    },
    stage: null,
    createdAt: step3.createdAt,
    updatedAt: expect.any(String),
    startTime: expect.any(String),
    endTime: expect.any(String),
  });

  // Check the final step updated the db as expected
  thing = await models.thing.findOne({
    id: thing!.id,
  });
  expect(thing!.name).toBe("Keelson updated");
  expect(thing!.age).toBe(32);
});

test("flows - text input validation", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  let { status, body } = await startFlow({
    name: "ValidationText",
    token,
    body: {},
  });

  expect(status).toBe(200);
  expect(body.steps[0].status).toBe("PENDING");

  const runId = body.id;
  let stepId = body.steps[0].id;

  ({ status, body } = await putStepValues({
    name: "ValidationText",
    runId,
    stepId,
    token,
    values: {
      postcode: "blah blah blah",
    },
  }));

  expect(status).toBe(200);
  expect(body.steps[0].ui).toEqual({
    __type: "ui.page",
    content: [
      {
        __type: "ui.input.text",
        disabled: false,
        label: "Postcode",
        name: "postcode",
        optional: false,
        placeholder: "e.g. N1 ABC",
        validationError: "not a valid postcode",
      },
    ],
    title: "Your postcode",
  });

  ({ status, body } = await putStepValues({
    name: "ValidationText",
    runId,
    stepId,
    token,
    values: {
      postcode: "E4 6ED",
    },
  }));

  expect(status).toBe(200);
  expect(body.steps[0].status).toBe("COMPLETED");
  expect(body.steps[0].value).toEqual({
    postcode: "E4 6ED",
  });
});

test("flows - boolean input validation", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  let { status, body } = await startFlow({
    name: "ValidationBoolean",
    token,
    body: {},
  });

  expect(status).toBe(200);
  expect(body.steps[0].status).toBe("PENDING");

  const runId = body.id;
  let stepId = body.steps[0].id;

  ({ status, body } = await putStepValues({
    name: "ValidationBoolean",
    runId,
    stepId,
    token,
    values: {
      good: false,
    },
  }));

  expect(status).toBe(200);
  expect(body.steps[0].ui).toEqual({
    __type: "ui.page",
    content: [
      {
        __type: "ui.input.boolean",
        disabled: false,
        mode: "checkbox",
        label: "Is it good?",
        name: "good",
        validationError: "it must be good",
        optional: false,
      },
    ],
    title: "Important question",
  });

  ({ status, body } = await putStepValues({
    name: "ValidationBoolean",
    runId,
    stepId,
    token,
    values: {
      good: true,
    },
  }));

  expect(status).toBe(200);
  expect(body.steps[0].status).toBe("COMPLETED");
  expect(body.steps[0].value).toEqual({
    good: true,
  });
});

test("flows - all inputs", async () => {
  const fileContents = "hello";
  const dataUrl = `data:text/plain;name=my-file.txt;base64,${Buffer.from(
    fileContents
  ).toString("base64")}`;

  const token = await getToken({ email: "admin@keel.xyz" });
  const res = await startFlow({
    name: "AllInputs",
    token,
    body: {
      text: "text",
      number: 1,
      file: dataUrl,
      date: "2021-01-01",
      timestamp: "2021-01-01T12:30:15.000Z",
      duration: "PT1000S",
      bool: true,
      decimal: 1.1,
      enum: MyEnum.Value1,
      markdown: "**Hello**",
    },
  });

  expect(res.status).toBe(200);
  expect(res.body).toEqual({
    config: {
      title: "All inputs",
    },
    createdAt: expect.any(String),
    id: expect.any(String),
    input: {
      date: "2021-01-01",
      duration: "PT1000S",
      file: "data:text/plain;name=my-file.txt;base64,aGVsbG8=",
      number: 1,
      text: "text",
      timestamp: "2021-01-01T12:30:15.000Z",
      bool: true,
      decimal: 1.1,
      enum: MyEnum.Value1,
      markdown: "**Hello**",
    },
    name: "AllInputs",
    startedBy: expect.any(String),
    status: "FAILED",
    steps: [],
    traceId: expect.any(String),
    updatedAt: expect.any(String),
  });
});

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

  // We are expecting 3 steps
  expect(flow).toEqual({
    id: res.body.id,
    traceId: res.body.traceId,
    status: "FAILED",
    name: "ErrorInStep",
    startedBy: expect.any(String),
    input: {},
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
    config: null,
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
    steps: [
      {
        id: res.body.steps[0].id,
        name: "timeout step",
        runId: res.body.id,
        stage: null,
        status: "FAILED",
        type: "FUNCTION",
        value: null,
        error: "Step function timed out after 1ms",
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
        error: "Step function timed out after 1ms",
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
        error: "Step function timed out after 1ms",
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
        error: "Step function timed out after 1ms",
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
        error: "Step function timed out after 1ms",
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        startTime: expect.any(String),
        endTime: expect.any(String),
        ui: null,
      },
    ],
    createdAt: res.body.createdAt,
    updatedAt: expect.any(String),
    config: null,
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
    config: null,
  });
});

test("flows - duplicate step name and UI name", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  const res = await startFlow({
    name: "DuplicateStepAndUiName",
    token,
    body: {},
  });
  expect(res.status).toBe(200);
  expect(res.body).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "RUNNING",
    name: "DuplicateStepAndUiName",
    startedBy: expect.any(String),
    input: {},
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
      title: "Duplicate step and ui name",
    },
  });

  const flow = await untilFlowFinished({
    name: "DuplicateStepAndUiName",
    id: res.body.id,
    token,
  });

  expect(flow).toEqual({
    id: res.body.id,
    traceId: res.body.traceId,
    status: "FAILED",
    name: "DuplicateStepAndUiName",
    startedBy: expect.any(String),
    input: {},
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
    config: null,
  });
});

test("flows - myRuns", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  const res = await startFlow({ name: "ErrorInFlow", token, body: {} });
  expect(res.status).toBe(200);

  let { status, body } = await startFlow({
    name: "scalarStep",
    token,
    body: {},
  });
  expect(status).toEqual(200);

  await untilFlowFinished({
    name: "scalarStep",
    id: body.id,
    token,
  });

  let resListRuns = await listMyRuns({
    token: token,
    params: { status: "FAILED" },
  });
  expect(resListRuns.status).toBe(200);
  expect(resListRuns.body.length).toBe(1);

  resListRuns = await listMyRuns({
    token: token,
    params: { status: ["FAILED", "COMPLETED"] },
  });
  expect(resListRuns.status).toBe(200);
  expect(resListRuns.body.length).toBe(2);
});

test("flows - authorised starting, getting and listing flows", async () => {
  const adminToken = await getToken({ email: "admin@keel.xyz" });
  const userToken = await getToken({ email: "user@gmail.com" });

  const resListAdmin = await listFlows({ token: adminToken });
  expect(resListAdmin.status).toBe(200);
  expect(resListAdmin.body.flows.length).toBe(15);
  expect(resListAdmin.body.flows[0].name).toBe("ScalarStep");
  expect(resListAdmin.body.flows[1].name).toBe("MixedStepTypes");
  expect(resListAdmin.body.flows[2].name).toBe("Stepless");
  expect(resListAdmin.body.flows[3].name).toBe("SingleStep");
  expect(resListAdmin.body.flows[4].name).toBe("ErrorInStep");
  expect(resListAdmin.body.flows[5].name).toBe("ErrorInFlow");
  expect(resListAdmin.body.flows[6].name).toBe("TimeoutStep");
  expect(resListAdmin.body.flows[7].name).toBe("OnlyPages");
  expect(resListAdmin.body.flows[8].name).toBe("OnlyFunctions");
  expect(resListAdmin.body.flows[9].name).toBe("ValidationText");
  expect(resListAdmin.body.flows[10].name).toBe("ValidationBoolean");
  expect(resListAdmin.body.flows[11].name).toBe("AllInputs");
  expect(resListAdmin.body.flows[12].name).toBe("DuplicateStepName");
  expect(resListAdmin.body.flows[13].name).toBe("DuplicateStepAndUiName");
  expect(resListAdmin.body.flows[14].name).toBe("EnvStep");

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
  const resGet = await listFlows({ token: null });
  expect(resGet.status).toBe(401);
});

test("flows - env step", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  let { status, body } = await startFlow({
    name: "envStep",
    token,
    body: {},
  });
  expect(status).toEqual(200);

  const flow = await untilFlowFinished({
    name: "envStep",
    id: body.id,
    token,
  });

  expect(flow).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "COMPLETED",
    name: "EnvStep",
    input: {},
    startedBy: expect.any(String),
    steps: [
      {
        id: expect.any(String),
        name: "env step",
        runId: expect.any(String),
        stage: null,
        status: "COMPLETED",
        type: "FUNCTION",
        value: "Pedro",
        error: null,
        startTime: expect.any(String),
        endTime: expect.any(String),
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        ui: null,
      },
    ],
    createdAt: expect.any(String),
    updatedAt: expect.any(String),
    config: null,
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

async function listMyRuns({ token, params }) {
  const queryString = new URLSearchParams(params).toString();
  const url = `${process.env.KEEL_TESTING_API_URL}/flows/json/myRuns?${queryString}`;

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
  const timeout = 5000; // We'll wait up to 5 seconds

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
