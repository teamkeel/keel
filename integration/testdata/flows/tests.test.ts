import { resetDatabase, models } from "@teamkeel/testing";
import { MyEnum } from "@teamkeel/sdk";
import { beforeEach, expect, test } from "vitest";

beforeEach(resetDatabase);

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
    data: null,
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
    config: {
      title: "Scalar step",
    },
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
    data: null,
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
    data: null,
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
    data: null,
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
          hasValidationErrors: false,
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
    action: null,
  }));
  expect(status).toEqual(200);
  expect(body).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "AWAITING_INPUT",
    name: "OnlyPages",
    startedBy: expect.any(String),
    input: {},
    data: null,
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
          hasValidationErrors: false,
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
    action: null,
  }));
  expect(status).toEqual(200);
  expect(body).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "COMPLETED",
    name: "OnlyPages",
    startedBy: expect.any(String),
    input: {},
    data: null,
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
    data: null,
    name: "Stepless",
    startedBy: expect.any(String),
    status: "COMPLETED",
    steps: [],
    config: {
      title: "Stepless",
    },
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
    data: null,
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
    data: null,
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
    config: {
      title: "Single step",
    },
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
    data: null,
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
    data: null,
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
          hasValidationErrors: false,
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
    action: null,
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
    data: null,
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
    action: null,
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
    hasValidationErrors: true,
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
    action: null,
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
    action: null,
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
    hasValidationErrors: true,
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
    action: null,
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
    data: null,
    name: "AllInputs",
    startedBy: expect.any(String),
    status: "FAILED",
    steps: [],
    traceId: expect.any(String),
    updatedAt: expect.any(String),
  });
});

test("flows - with completion", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  const res = await startFlow({ name: "WithCompletion", token, body: {} });
  expect(res.status).toBe(200);

  expect(res.body).toEqual({
    id: res.body.id,
    traceId: res.body.traceId,
    status: "RUNNING",
    name: "WithCompletion",
    startedBy: expect.any(String),
    input: {},
    steps: [
      {
        id: res.body.steps[0].id,
        name: "my step",
        runId: res.body.id,
        stage: "starting",
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
    createdAt: res.body.createdAt,
    updatedAt: expect.any(String),
    data: null,
    config: {
      stages: [
        {
          description: "this is the starting stage",
          key: "starting",
          name: "Starting",
        },
        {
          description: "this is the ending stage",
          key: "ending",
          name: "Ending",
        },
      ],
      title: "With completion",
    },
  });

  const flow = await untilFlowFinished({
    name: "WithCompletion",
    id: res.body.id,
    token,
  });

  expect(flow).toEqual({
    id: res.body.id,
    traceId: res.body.traceId,
    status: "COMPLETED",
    name: "WithCompletion",
    startedBy: expect.any(String),
    input: {},
    steps: [
      {
        id: res.body.steps[0].id,
        name: "my step",
        runId: res.body.id,
        stage: "starting",
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
        name: "",
        runId: res.body.id,
        stage: "ending",
        status: "COMPLETED",
        type: "COMPLETE",
        value: null,
        error: null,
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        startTime: expect.any(String),
        endTime: expect.any(String),
        ui: {
          __type: "ui.complete",
          title: "Completed flow",
          description: "this complete page replaces the normal end page",
          stage: "ending",
          content: [
            { __type: "ui.display.markdown", content: "congratulations" },
          ],
        },
      },
    ],
    createdAt: res.body.createdAt,
    updatedAt: expect.any(String),
    data: {
      value: "flow value",
    },
    config: {
      stages: [
        {
          description: "this is the starting stage",
          key: "starting",
          name: "Starting",
        },
        {
          description: "this is the ending stage",
          key: "ending",
          name: "Ending",
        },
      ],
      title: "With completion",
    },
  });
});

test("flows - with completion - no contents and no returns", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  const res = await startFlow({
    name: "WithCompletionMinimal",
    token,
    body: {},
  });
  expect(res.status).toBe(200);

  expect(res.body).toEqual({
    id: res.body.id,
    traceId: res.body.traceId,
    status: "COMPLETED",
    name: "WithCompletionMinimal",
    startedBy: expect.any(String),
    input: {},
    steps: [
      {
        id: res.body.steps[0].id,
        name: "",
        runId: res.body.id,
        stage: null,
        status: "COMPLETED",
        type: "COMPLETE",
        value: null,
        error: null,
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        startTime: expect.any(String),
        endTime: expect.any(String),
        ui: {
          __type: "ui.complete",
          title: "Completed flow",
          content: [],
        },
      },
    ],
    createdAt: res.body.createdAt,
    updatedAt: expect.any(String),
    data: null,
    config: {
      title: "With completion minimal",
    },
  });

  const flow = await getFlowRun({
    name: "WithCompletionMinimal",
    id: res.body.id,
    token,
  });

  expect(flow.body).toEqual({
    id: res.body.id,
    traceId: res.body.traceId,
    status: "COMPLETED",
    name: "WithCompletionMinimal",
    startedBy: expect.any(String),
    input: {},
    steps: [
      {
        id: expect.any(String),
        name: "",
        runId: res.body.id,
        stage: null,
        status: "COMPLETED",
        type: "COMPLETE",
        value: null,
        error: null,
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
        startTime: expect.any(String),
        endTime: expect.any(String),
        ui: {
          __type: "ui.complete",
          title: "Completed flow",
          content: [],
        },
      },
    ],
    createdAt: res.body.createdAt,
    updatedAt: expect.any(String),
    data: null,
    config: {
      title: "With completion minimal",
    },
  });
});

test("flows - with returned data", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  const res = await startFlow({ name: "WithReturnedData", token, body: {} });
  expect(res.status).toBe(200);

  expect(res.body).toEqual({
    id: res.body.id,
    traceId: res.body.traceId,
    status: "RUNNING",
    name: "WithReturnedData",
    startedBy: expect.any(String),
    input: {},
    steps: [
      {
        id: res.body.steps[0].id,
        name: "my step",
        runId: res.body.id,
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
    createdAt: res.body.createdAt,
    updatedAt: expect.any(String),
    data: null,
    config: {
      title: "With returned data",
    },
  });

  const flow = await untilFlowFinished({
    name: "WithReturnedData",
    id: res.body.id,
    token,
  });

  expect(flow).toEqual({
    id: res.body.id,
    traceId: res.body.traceId,
    status: "COMPLETED",
    name: "WithReturnedData",
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
    ],
    createdAt: res.body.createdAt,
    updatedAt: expect.any(String),
    data: "hello",
    config: {
      title: "With returned data",
    },
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

test("flows - authorised listing flows", async () => {
  const adminToken = await getToken({ email: "admin@keel.xyz" });
  const userToken = await getToken({ email: "user@gmail.com" });

  const identity = await models.identity.findOne({
    email: "user@gmail.com",
    issuer: "https://keel.so",
  });

  await models.user.create({ team: "myTeam", identityId: identity!.id });

  const resListAdmin = await listFlows({ token: adminToken });
  expect(resListAdmin.body.flows.length).toBe(16);
  expect(resListAdmin.body.flows[0].name).toBe("ScalarStep");
  expect(resListAdmin.body.flows[1].name).toBe("MixedStepTypes");
  expect(resListAdmin.body.flows[2].name).toBe("Stepless");
  expect(resListAdmin.body.flows[3].name).toBe("SingleStep");
  expect(resListAdmin.body.flows[4].name).toBe("ErrorInFlow");
  expect(resListAdmin.body.flows[5].name).toBe("OnlyPages");
  expect(resListAdmin.body.flows[6].name).toBe("OnlyFunctions");
  expect(resListAdmin.body.flows[7].name).toBe("ValidationText");
  expect(resListAdmin.body.flows[8].name).toBe("ValidationBoolean");
  expect(resListAdmin.body.flows[9].name).toBe("AllInputs");
  expect(resListAdmin.body.flows[10].name).toBe("EnvStep");
  expect(resListAdmin.body.flows[11].name).toBe("MultipleActions");
  expect(resListAdmin.body.flows[12].name).toBe("WithCompletion");
  expect(resListAdmin.body.flows[13].name).toBe("WithCompletionMinimal");
  expect(resListAdmin.body.flows[14].name).toBe("WithReturnedData");
  expect(resListAdmin.body.flows[15].name).toBe("ExpressionPermissionIsTrue");

  const resListUser = await listFlows({ token: userToken });
  expect(resListUser.status).toBe(200);
  expect(resListUser.body.flows.length).toBe(4);
  expect(resListUser.body.flows[0].name).toBe("UserFlow");
  expect(resListUser.body.flows[1].name).toBe("ExpressionPermissionCtx");
  expect(resListUser.body.flows[2].name).toBe("ExpressionPermissionEnv");
  expect(resListUser.body.flows[3].name).toBe("ExpressionPermissionIsTrue");
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

test("flows - authorised starting flow with true expression", async () => {
  const token = await getToken({ email: "user@gmail.com" });
  const res = await startFlow({
    name: "expressionPermissionIsTrue",
    token,
    body: {},
  });
  expect(res.status).toBe(200);
});

test("flows - not authorised starting flow with backlink expression", async () => {
  const token = await getToken({ email: "user@gmail.com" });
  const res = await startFlow({
    name: "ExpressionPermissionCtx",
    token,
    body: {},
  });
  expect(res.status).toBe(403);
});

test("flows - unauthorised (wrong team) starting flow with backlink expression", async () => {
  const token = await getToken({ email: "user@keel.xyz" });

  const identity = await models.identity.findOne({
    email: "user@keel.xyz",
    issuer: "https://keel.so",
  });

  const user = await models.user.create({
    team: "wrongTeam",
    identityId: identity!.id,
  });

  const res = await startFlow({
    name: "ExpressionPermissionCtx",
    token: token,
    body: {},
  });
  expect(res.status).toBe(403);
});

test("flows - authorised starting flow with backlink expression", async () => {
  const token = await getToken({ email: "user@keel.xyz" });

  const identity = await models.identity.findOne({
    email: "user@keel.xyz",
    issuer: "https://keel.so",
  });

  const user = await models.user.create({
    team: "myTeam",
    identityId: identity!.id,
  });

  const res = await startFlow({
    name: "ExpressionPermissionCtx",
    token: token,
    body: {},
  });
  expect(res.status).toBe(200);
});

test("flows - authorised starting flow with env var expression", async () => {
  const token = await getToken({ email: "user@keel.xyz" });

  const identity = await models.identity.findOne({
    email: "user@keel.xyz",
    issuer: "https://keel.so",
  });

  const user = await models.user.create({
    team: "myTeam",
    identityId: identity!.id,
  });

  const res = await startFlow({
    name: "ExpressionPermissionEnv",
    token: token,
    body: {},
  });
  expect(res.status).toBe(200);
});

test("flows - unauthorised (wrong team) starting flow with env var expression", async () => {
  const token = await getToken({ email: "user@keel.xyz" });

  const identity = await models.identity.findOne({
    email: "user@keel.xyz",
    issuer: "https://keel.so",
  });

  const user = await models.user.create({
    team: "wrongTeam",
    identityId: identity!.id,
  });

  const res = await startFlow({
    name: "ExpressionPermissionEnv",
    token: token,
    body: {},
  });
  expect(res.status).toBe(403);
});

test("flows - unauthenticated starting flow with backlink expression", async () => {
  const res = await startFlow({
    name: "ExpressionPermissionCtx",
    token: null,
    body: {},
  });
  expect(res.status).toBe(401);
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
    data: null,
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
      {
        id: expect.any(String),
        name: "identity step",
        runId: expect.any(String),
        stage: null,
        status: "COMPLETED",
        type: "FUNCTION",
        value: "admin@keel.xyz",
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
      title: "Env step",
    },
  });
});

test("flows - multiple actions - finish", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  let { status, body } = await startFlow({
    name: "MultipleActions",
    token,
    body: {},
  });
  expect(status).toEqual(200);
  expect(body).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "AWAITING_INPUT",
    name: "MultipleActions",
    startedBy: expect.any(String),
    input: {},
    data: null,
    steps: [
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
          actions: [
            {
              label: "finish",
              mode: "primary",
              value: "finish",
            },
            {
              label: "continue",
              mode: "primary",
              value: "continue",
            },
          ],
          content: [
            {
              __type: "ui.input.boolean",
              disabled: false,
              label: "Did you like the things?",
              mode: "checkbox",
              name: "yesno",
              optional: false,
            },
          ],
          hasValidationErrors: false,
          title: "Continue flow?",
        },
      },
    ],
    createdAt: expect.any(String),
    updatedAt: expect.any(String),
    config: {
      title: "Multiple actions",
    },
  });

  ({ status, body } = await putStepValues({
    name: "MultipleActions",
    runId: body.id,
    stepId: body.steps[0].id,
    token,
    values: { yesno: true },
    action: "finish",
  }));
  expect(status).toEqual(200);
  expect(body).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "COMPLETED",
    name: "MultipleActions",
    startedBy: expect.any(String),
    input: {},
    data: null,
    steps: expect.any(Array),
    createdAt: expect.any(String),
    updatedAt: expect.any(String),
    config: {
      title: "Multiple actions",
    },
  });
});

test("flows - multiple actions - continue", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  let { status, body } = await startFlow({
    name: "MultipleActions",
    token,
    body: {},
  });
  expect(status).toEqual(200);
  expect(body).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "AWAITING_INPUT",
    name: "MultipleActions",
    startedBy: expect.any(String),
    input: {},
    data: null,
    steps: [
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
          actions: [
            {
              label: "finish",
              mode: "primary",
              value: "finish",
            },
            {
              label: "continue",
              mode: "primary",
              value: "continue",
            },
          ],
          content: [
            {
              __type: "ui.input.boolean",
              disabled: false,
              label: "Did you like the things?",
              mode: "checkbox",
              name: "yesno",
              optional: false,
            },
          ],
          hasValidationErrors: false,
          title: "Continue flow?",
        },
      },
    ],
    createdAt: expect.any(String),
    updatedAt: expect.any(String),
    config: {
      title: "Multiple actions",
    },
  });

  // Provide the values for the pending UI step
  ({ status, body } = await putStepValues({
    name: "MultipleActions",
    runId: body.id,
    stepId: body.steps[0].id,
    token,
    values: {},
    action: "continue",
  }));
  expect(status).toEqual(200);
  expect(body).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "AWAITING_INPUT",
    name: "MultipleActions",
    startedBy: expect.any(String),
    input: {},
    data: null,
    steps: [
      {
        id: expect.any(String),
        name: "question",
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
        name: "another-question",
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
              __type: "ui.input.text",
              label: "Name",
              disabled: false,
              optional: false,
              name: "name",
            },
          ],
          hasValidationErrors: false,
          title: "Another question",
        },
      },
    ],
    createdAt: expect.any(String),
    updatedAt: expect.any(String),
    config: {
      title: "Multiple actions",
    },
  });

  ({ status, body } = await putStepValues({
    name: "MixedStepTypes",
    runId: body.id,
    stepId: body.steps[1].id,
    token,
    values: { yesno: true },
    action: null,
  }));

  expect(status).toEqual(200);
  expect(body).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "COMPLETED",
    name: "MultipleActions",
    startedBy: expect.any(String),
    input: {},
    data: null,
    steps: [
      {
        id: expect.any(String),
        name: "question",
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
        name: "another-question",
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
      title: "Multiple actions",
    },
  });
});

test("flows - cancelling - with pending ui step", async () => {
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
    data: null,
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
          hasValidationErrors: false,
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

  const resp = await cancelFlow({ name: "MixedStepTypes", runId, token });
  expect(resp.status).toBe(200);
  expect(resp.body.status).toBe("CANCELLED"); // Flow run's status is CANCELLED
  expect(resp.body.steps[1]).toEqual({
    id: expect.any(String),
    runId: runId,
    stage: null,
    name: "confirm thing",
    error: null,
    ui: null, // We do not have the full UI config because this step is now cancelled
    status: "CANCELLED", // This step is now cancelled
    type: "UI",
    value: null,
    createdAt: expect.any(String),
    updatedAt: expect.any(String),
    startTime: expect.any(String),
    endTime: null,
  });
});

test("flows - multiple actions - invalid action", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  let { status, body } = await startFlow({
    name: "MultipleActions",
    token,
    body: {},
  });
  expect(status).toEqual(200);
  expect(body).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "AWAITING_INPUT",
    name: "MultipleActions",
    startedBy: expect.any(String),
    input: {},
    data: null,
    steps: [
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
          actions: [
            {
              label: "finish",
              mode: "primary",
              value: "finish",
            },
            {
              label: "continue",
              mode: "primary",
              value: "continue",
            },
          ],
          content: [
            {
              __type: "ui.input.boolean",
              disabled: false,
              label: "Did you like the things?",
              mode: "checkbox",
              name: "yesno",
              optional: false,
            },
          ],
          hasValidationErrors: false,
          title: "Continue flow?",
        },
      },
    ],
    createdAt: expect.any(String),
    updatedAt: expect.any(String),
    config: {
      title: "Multiple actions",
    },
  });

  // Provide the values for the pending UI step
  ({ status, body } = await putStepValues({
    name: "MultipleActions",
    runId: body.id,
    stepId: body.steps[0].id,
    token,
    values: {},
    action: "invalid",
  }));
  expect(status).toEqual(200);
  expect(body).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "AWAITING_INPUT",
    name: "MultipleActions",
    startedBy: expect.any(String),
    input: {},
    data: null,
    steps: [
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
          actions: [
            {
              label: "finish",
              mode: "primary",
              value: "finish",
            },
            {
              label: "continue",
              mode: "primary",
              value: "continue",
            },
          ],
          content: [
            {
              __type: "ui.input.boolean",
              disabled: false,
              label: "Did you like the things?",
              mode: "checkbox",
              name: "yesno",
              optional: false,
            },
          ],
          hasValidationErrors: true,
          title: "Continue flow?",
          validationError: "invalid action",
        },
      },
    ],
    createdAt: expect.any(String),
    updatedAt: expect.any(String),
    config: {
      title: "Multiple actions",
    },
  });
});

test("flows - stats", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  await startFlow({ name: "ErrorInFlow", token, body: {} });

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

  let stats = await listStats({
    token: token,
    params: { interval: "daily" },
  });

  expect(stats.status).toBe(200);
  expect(stats.body).toEqual([
    {
      activeRuns: 0,
      completedToday: 0,
      errorRate: 1,
      lastRun: expect.any(String),
      name: "ErrorInFlow",
      timeSeries: [
        {
          failedRuns: 1,
          time: expect.any(String),
          totalRuns: 1,
        },
      ],
      totalRuns: 1,
    },
    {
      activeRuns: 0,
      completedToday: 1,
      errorRate: 0,
      lastRun: expect.any(String),
      name: "ScalarStep",
      timeSeries: [
        {
          failedRuns: 0,
          time: expect.any(String),
          totalRuns: 1,
        },
      ],
      totalRuns: 1,
    },
  ]);
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

async function listStats({ token, params }) {
  const queryString = new URLSearchParams(params).toString();
  const url = `${process.env.KEEL_TESTING_API_URL}/flows/json/stats?${queryString}`;

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

async function cancelFlow({ name, runId, token }) {
  const res = await fetch(
    `${process.env.KEEL_TESTING_API_URL}/flows/json/${name}/${runId}/cancel`,
    {
      method: "POST",
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
