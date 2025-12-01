import { resetDatabase, models, flows } from "@teamkeel/testing";
import { MyEnum, Duration } from "@teamkeel/sdk";
import { beforeEach, expect, test } from "vitest";

beforeEach(resetDatabase);

test("flows - scalar step", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const f = await flows.scalarStep.withAuthToken(token).start({});

  const flow = await flows.scalarStep.withAuthToken(token).untilFinished(f.id);

  expect(flow).toEqual({
    id: expect.any(String),
    status: "COMPLETED",
    name: "ScalarStep",
    traceId: expect.any(String),
    input: {},
    error: null,
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
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        ui: null,
      },
    ],
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "Scalar step",
    },
  });
});

test("flows - map step", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const f = await flows.mapStep.withAuthToken(token).start({});

  const flow = await flows.mapStep.withAuthToken(token).untilFinished(f.id);

  expect(flow).toEqual({
    id: expect.any(String),
    status: "COMPLETED",
    name: "MapStep",
    traceId: expect.any(String),
    input: {},
    error: null,
    data: null,
    startedBy: expect.any(String),
    steps: [
      {
        id: expect.any(String),
        name: "create map",
        runId: expect.any(String),
        stage: null,
        status: "COMPLETED",
        type: "FUNCTION",
        value: {
          name: "Keelson",
          age: 25,
          active: true,
          nested: { city: "London", country: "UK" },
        },
        error: null,
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "verify map object",
        runId: expect.any(String),
        stage: null,
        status: "COMPLETED",
        type: "FUNCTION",
        value: {
          hasName: true,
          hasAge: true,
          hasActive: true,
          hasNested: true,
          // Map is converted to a plain object during serialization
          isMap: false,
          isObject: true,
        },
        error: null,
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        ui: null,
      },
    ],
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "Map step",
    },
  });
});

test("flows - date step", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const f = await flows.dateStep.withAuthToken(token).start({});

  const flow = await flows.dateStep.withAuthToken(token).untilFinished(f.id);

  expect(flow).toEqual({
    id: expect.any(String),
    status: "COMPLETED",
    name: "DateStep",
    traceId: expect.any(String),
    input: {},
    error: null,
    data: null,
    startedBy: expect.any(String),
    steps: [
      {
        id: expect.any(String),
        name: "create date",
        runId: expect.any(String),
        stage: null,
        status: "COMPLETED",
        type: "FUNCTION",
        value: new Date("2024-01-15T10:30:00.000Z"),
        error: null,
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "verify date object",
        runId: expect.any(String),
        stage: null,
        status: "COMPLETED",
        type: "FUNCTION",
        value: {
          isDate: true,
          // isoString gets deserialized back to a Date because it matches ISO pattern
          isoString: new Date("2024-01-15T10:30:00.000Z"),
          timestamp: expect.any(Number),
        },
        error: null,
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        ui: null,
      },
    ],
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "Date step",
    },
  });
});

test("flows - model step", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const f = await flows.modelStep.withAuthToken(token).start({
    name: "Test Thing",
    age: 42,
  });

  const flow = await flows.modelStep.withAuthToken(token).untilFinished(f.id);

  expect(flow).toEqual({
    id: expect.any(String),
    status: "COMPLETED",
    name: "ModelStep",
    traceId: expect.any(String),
    input: {
      name: "Test Thing",
      age: 42,
    },
    error: null,
    data: null,
    startedBy: expect.any(String),
    steps: [
      {
        id: expect.any(String),
        name: "create and return model",
        runId: expect.any(String),
        stage: null,
        status: "COMPLETED",
        type: "FUNCTION",
        value: {
          id: expect.any(String),
          name: "Test Thing",
          age: 42,
          createdAt: expect.any(Date),
          updatedAt: expect.any(Date),
        },
        error: null,
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "verify model",
        runId: expect.any(String),
        stage: null,
        status: "COMPLETED",
        type: "FUNCTION",
        value: {
          hasId: true,
          hasName: true,
          hasAge: true,
          createdAtIsDate: true,
          updatedAtIsDate: true,
          canCallDateMethod: true,
        },
        error: null,
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        ui: null,
      },
    ],
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "Model step",
    },
  });
});

test("flows - file step", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const f = await flows.fileStep.withAuthToken(token).start({});

  const flow = await flows.fileStep.withAuthToken(token).untilFinished(f.id);

  expect(flow).toEqual({
    id: expect.any(String),
    status: "COMPLETED",
    name: "FileStep",
    traceId: expect.any(String),
    input: {},
    error: null,
    data: null,
    startedBy: expect.any(String),
    steps: [
      {
        id: expect.any(String),
        name: "create and store file",
        runId: expect.any(String),
        stage: null,
        status: "COMPLETED",
        type: "FUNCTION",
        value: {
          __keel_type: "file",
          key: expect.any(String),
          filename: "test-file.txt",
          contentType: "text/plain",
          size: expect.any(Number),
        },
        error: null,
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "verify file object",
        runId: expect.any(String),
        stage: null,
        status: "COMPLETED",
        type: "FUNCTION",
        value: {
          hasFilename: true,
          hasContentType: true,
          hasKey: true,
          hasSize: true,
          canRead: true,
          isFileInstance: true,
        },
        error: null,
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        ui: null,
      },
    ],
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "File step",
    },
  });
});

test("flows - only functions with config", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const f = await flows.onlyFunctions.withAuthToken(token).start({
    name: "My Thing",
    age: 25,
  });

  expect(f).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "RUNNING",
    name: "OnlyFunctions",
    startedBy: expect.any(String),
    input: {
      name: "My Thing",
      age: 25,
    },
    error: null,
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

  const flow = await flows.onlyFunctions
    .withAuthToken(token)
    .untilFinished(f.id);

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
    error: null,
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
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
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
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        ui: null,
      },
    ],
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
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

  const f = await flows.onlyPages.withAuthToken(token).start({});

  expect(f).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "AWAITING_INPUT",
    name: "OnlyPages",
    startedBy: expect.any(String),
    input: {},
    error: null,
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
        startTime: expect.any(Date),
        endTime: null,
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
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
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "Only pages",
    },
  });

  // Provide the values for the pending UI step
  const updatedFlow = await flows.onlyPages
    .withAuthToken(token)
    .putStepValues(f.id, f.steps[0].id, {});

  expect(updatedFlow).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "AWAITING_INPUT",
    name: "OnlyPages",
    startedBy: expect.any(String),
    input: {},
    error: null,
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
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
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
        startTime: expect.any(Date),
        endTime: null,
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        allowBack: true,
        ui: {
          __type: "ui.page",
          allowBack: true,
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
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "Only pages",
    },
  });

  const finalFlow = await flows.onlyPages
    .withAuthToken(token)
    .putStepValues(updatedFlow.id, updatedFlow.steps[1].id, { yesno: true });

  expect(finalFlow).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "COMPLETED",
    name: "OnlyPages",
    startedBy: expect.any(String),
    input: {},
    error: null,
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
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
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
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        ui: null,
      },
    ],
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "Only pages",
    },
  });
});

test("flows - back on pages", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const f = await flows.onlyPages.withAuthToken(token).start({});

  expect(f).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "AWAITING_INPUT",
    name: "OnlyPages",
    startedBy: expect.any(String),
    input: {},
    error: null,
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
        startTime: expect.any(Date),
        endTime: null,
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
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
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "Only pages",
    },
  });

  // Provide the values for the pending UI step
  const updatedFlow = await flows.onlyPages
    .withAuthToken(token)
    .putStepValues(f.id, f.steps[0].id, {});

  expect(updatedFlow).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "AWAITING_INPUT",
    name: "OnlyPages",
    startedBy: expect.any(String),
    input: {},
    error: null,
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
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
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
        startTime: expect.any(Date),
        endTime: null,
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        allowBack: true,
        ui: {
          __type: "ui.page",
          allowBack: true,
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
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "Only pages",
    },
  });

  // go back one step
  const backFlow = await flows.onlyPages
    .withAuthToken(token)
    .back(updatedFlow.id);
  expect(backFlow).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "AWAITING_INPUT",
    name: "OnlyPages",
    startedBy: expect.any(String),
    input: {},
    error: null,
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
        startTime: expect.any(Date),
        endTime: null,
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
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
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "Only pages",
    },
  });

  // Provide the values for the pending UI step
  const updatedAgainFlow = await flows.onlyPages
    .withAuthToken(token)
    .putStepValues(f.id, f.steps[0].id, {});

  expect(updatedAgainFlow).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "AWAITING_INPUT",
    name: "OnlyPages",
    startedBy: expect.any(String),
    input: {},
    error: null,
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
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
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
        startTime: expect.any(Date),
        endTime: null,
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        allowBack: true,
        ui: {
          __type: "ui.page",
          allowBack: true,
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
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "Only pages",
    },
  });

  const finalFlow = await flows.onlyPages
    .withAuthToken(token)
    .putStepValues(updatedFlow.id, updatedFlow.steps[1].id, { yesno: true });

  expect(finalFlow).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "COMPLETED",
    name: "OnlyPages",
    startedBy: expect.any(String),
    input: {},
    error: null,
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
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
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
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        ui: null,
      },
    ],
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "Only pages",
    },
  });
});

test("flows - stepless flow", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const f = await flows.stepless.withAuthToken(token).start({});

  const flow = await flows.stepless.withAuthToken(token).untilFinished(f.id);

  // Flow has no steps so should be synchronously completed
  expect(flow).toEqual({
    id: f.id,
    input: {},
    error: null,
    data: null,
    name: "Stepless",
    startedBy: expect.any(String),
    status: "COMPLETED",
    steps: [],
    config: {
      title: "Stepless",
    },
    traceId: expect.any(String),
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
  });

  const things = await models.thing.findMany();
  expect(things.length).toBe(1);
  expect(things[0].name).toBe("Keelson");
});

test("flows - first step is a function", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const f = await flows.singleStep.withAuthToken(token).start({});

  // First step is a function so should be in status NEW - it will get run async via the queue
  expect(f).toEqual({
    id: expect.any(String),
    input: {},
    name: "SingleStep",
    startedBy: expect.any(String),
    status: "RUNNING",
    error: null,
    data: null,
    steps: [
      {
        id: expect.any(String),
        runId: f.id,
        stage: null,
        name: "insert thing",
        error: null,
        ui: null,
        status: "NEW",
        type: "FUNCTION",
        value: null,
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: null,
        endTime: null,
      },
    ],
    config: {
      title: "Single step",
    },
    traceId: expect.any(String),
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
  });

  const flow = await flows.singleStep.withAuthToken(token).untilFinished(f.id);

  // Now the flow has finished the run and step statuses should have been
  // updated and the returned value stored against the step
  expect(flow).toEqual({
    id: f.id,
    input: {},
    name: "SingleStep",
    startedBy: expect.any(String),
    status: "COMPLETED",
    error: null,
    data: null,
    steps: [
      {
        id: f.steps[0].id,
        runId: f.id,
        stage: null,
        name: "insert thing",
        error: null,
        ui: null,
        status: "COMPLETED",
        type: "FUNCTION",
        value: {
          number: 10,
        },
        createdAt: f.steps[0].createdAt,
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
      },
    ],
    config: {
      title: "Single step",
    },
    traceId: f.traceId,
    createdAt: f.createdAt,
    updatedAt: expect.any(Date),
  });
});

test("flows - alternating step types", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const f = await flows.mixedStepTypes.withAuthToken(token).start({
    name: "Keelson",
    age: 23,
  });

  // First step is a function so API response should show that as NEW
  expect(f).toEqual({
    id: expect.any(String),
    input: {
      name: "Keelson",
      age: 23,
    },
    error: null,
    data: null,
    name: "MixedStepTypes",
    startedBy: expect.any(String),
    status: "RUNNING",
    steps: [
      {
        id: expect.any(String),
        runId: f.id,
        stage: null,
        name: "insert thing",
        error: null,
        ui: null,
        status: "NEW",
        type: "FUNCTION",
        value: null,
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: null,
        endTime: null,
      },
    ],
    config: {
      title: "Mixed step types",
    },
    traceId: expect.any(String),
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
  });

  const runId = f.id;
  const traceId = f.traceId;
  let step1 = f.steps[0];

  // The second step is a page with UI so we wait until the flow has reached that point
  const body = await flows.mixedStepTypes
    .withAuthToken(token)
    .untilAwaitingInput(runId);
  expect(body).toEqual({
    id: runId,
    name: "MixedStepTypes",
    startedBy: expect.any(String),
    status: "AWAITING_INPUT", // Flow is now awaiting input
    input: {
      name: "Keelson",
      age: 23,
    },
    error: null,
    data: null,
    config: {
      title: "Mixed step types",
    },
    traceId,
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
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
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
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
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
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
  const updatedFlow = await flows.mixedStepTypes
    .withAuthToken(token)
    .putStepValues(runId, step2.id, {
      name: "Keelson updated",
      age: 32,
    });

  expect(updatedFlow).toEqual({
    id: runId,
    name: "MixedStepTypes",
    startedBy: expect.any(String),
    status: "RUNNING",
    input: {
      name: "Keelson",
      age: 23,
    },
    error: null,
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
        updatedAt: expect.any(Date),
        endTime: expect.any(Date),
      },
      {
        // The final step is now pending and will be run via the queue
        id: expect.any(String),
        runId,
        stage: null,
        status: "NEW",
        type: "FUNCTION",
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
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
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
  });

  step2 = updatedFlow.steps[1];
  let step3 = updatedFlow.steps[2];

  const finalFlow = await flows.mixedStepTypes
    .withAuthToken(token)
    .untilFinished(runId);
  expect(finalFlow.status).toBe("COMPLETED");
  expect(finalFlow.steps[2]).toEqual({
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
    updatedAt: expect.any(Date),
    startTime: expect.any(Date),
    endTime: expect.any(Date),
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
  let f = await flows.validationText.withAuthToken(token).start({});

  expect(f.steps[0].status).toBe("PENDING");

  const runId = f.id;
  let stepId = f.steps[0].id;

  f = await flows.validationText
    .withAuthToken(token)
    .putStepValues(runId, stepId, {
      postcode: "blah blah blah",
    });

  expect(f.steps[0].ui).toEqual({
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

  f = await flows.validationText
    .withAuthToken(token)
    .putStepValues(runId, stepId, {
      postcode: "E4 6ED",
    });

  expect(f.steps[0].status).toBe("COMPLETED");
  expect(f.steps[0].value).toEqual({
    postcode: "E4 6ED",
  });
});

test("flows - boolean input validation", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  let f = await flows.validationBoolean.withAuthToken(token).start({});

  expect(f.steps[0].status).toBe("PENDING");

  const runId = f.id;
  let stepId = f.steps[0].id;

  f = await flows.validationBoolean
    .withAuthToken(token)
    .putStepValues(runId, stepId, {
      good: false,
    });

  expect(f.steps[0].ui).toEqual({
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

  f = await flows.validationBoolean
    .withAuthToken(token)
    .putStepValues(runId, stepId, {
      good: true,
    });

  expect(f.steps[0].status).toBe("COMPLETED");
  expect(f.steps[0].value).toEqual({
    good: true,
  });
});

test("flows - page validation", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  let f = await flows.validationPage.withAuthToken(token).start({});

  expect(f.steps[0].status).toBe("PENDING");

  const runId = f.id;
  let stepId = f.steps[0].id;

  f = await flows.validationPage
    .withAuthToken(token)
    .putStepValues(runId, stepId, {});

  expect(f.steps[0].ui).toEqual({
    __type: "ui.page",
    content: [
      {
        __type: "ui.input.text",
        disabled: false,
        label: "Email",
        name: "email",
        optional: true,
      },
      {
        __type: "ui.input.text",
        disabled: false,
        label: "Phone",
        name: "phone",
        optional: true,
      },
    ],
    hasValidationErrors: true,
    validationError: "Email or phone is required",
  });

  f = await flows.validationPage
    .withAuthToken(token)
    .putStepValues(runId, stepId, {
      email: "keelson.keel.xyz",
    });

  expect(f.steps[0].ui).toEqual({
    __type: "ui.page",
    content: [
      {
        __type: "ui.input.text",
        disabled: false,
        label: "Email",
        name: "email",
        optional: true,
        validationError: "Not a valid email",
      },
      {
        __type: "ui.input.text",
        disabled: false,
        label: "Phone",
        name: "phone",
        optional: true,
      },
    ],
    hasValidationErrors: true,
  });

  f = await flows.validationPage
    .withAuthToken(token)
    .putStepValues(runId, stepId, {
      email: "keelson@keel.xyz",
    });

  expect(f.steps[0].status).toBe("COMPLETED");
  expect(f.steps[0].value).toEqual({
    email: "keelson@keel.xyz",
  });
});

test("flows - page validation with actions", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  let f = await flows.validationPageWithAction.withAuthToken(token).start({});

  expect(f.steps[0].status).toBe("PENDING");

  const runId = f.id;
  let stepId = f.steps[0].id;

  f = await flows.validationPageWithAction
    .withAuthToken(token)
    .putStepValues(runId, stepId, {}, "next");

  expect(f.steps[0].ui).toEqual({
    __type: "ui.page",
    content: [
      {
        __type: "ui.input.text",
        disabled: false,
        label: "Email",
        name: "email",
        optional: true,
      },
      {
        __type: "ui.input.text",
        disabled: false,
        label: "Phone",
        name: "phone",
        optional: true,
      },
    ],
    actions: [
      {
        label: "Cancel",
        mode: "primary",
        value: "cancel",
      },
      {
        label: "Next",
        mode: "primary",
        value: "next",
      },
    ],
    hasValidationErrors: true,
    validationError: "Email or phone is required",
  });

  f = await flows.validationPageWithAction.withAuthToken(token).putStepValues(
    runId,
    stepId,
    {
      email: "keelson.keel.xyz",
    },
    "next"
  );

  expect(f.steps[0].ui).toEqual({
    __type: "ui.page",
    content: [
      {
        __type: "ui.input.text",
        disabled: false,
        label: "Email",
        name: "email",
        optional: true,
        validationError: "Not a valid email",
      },
      {
        __type: "ui.input.text",
        disabled: false,
        label: "Phone",
        name: "phone",
        optional: true,
      },
    ],
    actions: [
      {
        label: "Cancel",
        mode: "primary",
        value: "cancel",
      },
      {
        label: "Next",
        mode: "primary",
        value: "next",
      },
    ],
    hasValidationErrors: true,
  });

  f = await flows.validationPageWithAction.withAuthToken(token).putStepValues(
    runId,
    stepId,
    {
      email: "keelson.keel.xyz",
    },
    "cancel"
  );

  expect(f.steps[0].status).toBe("COMPLETED");
  expect(f.steps[0].value).toEqual({
    email: "keelson.keel.xyz",
  });
});

test("flows - page validation with actions - invalid action", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  let f = await flows.validationPageWithAction.withAuthToken(token).start({});

  const runId = f.id;
  let stepId = f.steps[0].id;

  // Provide an invalid action that doesn't exist - this should fail the step
  f = await flows.validationPageWithAction
    .withAuthToken(token)
    .putStepValues(runId, stepId, { email: "test@example.com" }, "invalid");

  // The step should have failed with an error
  expect(f.steps[0].status).toBe("FAILED");
  expect(f.steps[0].error).toContain(
    'invalid action "invalid". Valid actions are: cancel, next'
  );
});

test("flows - page validation with actions - label instead of value", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  let f = await flows.validationPageWithAction.withAuthToken(token).start({});

  const runId = f.id;
  let stepId = f.steps[0].id;

  // Provide a label instead of a value (common mistake) - this should fail the step
  f = await flows.validationPageWithAction
    .withAuthToken(token)
    .putStepValues(runId, stepId, { email: "test@example.com" }, "Cancel");

  // The step should have failed with an error
  expect(f.steps[0].status).toBe("FAILED");
  expect(f.steps[0].error).toContain(
    'invalid action "Cancel". Valid actions are: cancel, next'
  );
});

test("flows - page validation with actions - missing action", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  let f = await flows.validationPageWithAction.withAuthToken(token).start({});

  const runId = f.id;
  let stepId = f.steps[0].id;

  // Don't provide an action when actions are defined (action is required) - this should fail the step
  f = await flows.validationPageWithAction
    .withAuthToken(token)
    .putStepValues(runId, stepId, { email: "test@example.com" });

  // The step should have failed with an error
  expect(f.steps[0].status).toBe("FAILED");
  expect(f.steps[0].error).toContain(
    "action is required. Valid actions are: cancel, next"
  );
});

test("flows - page validation - action provided when none defined", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  let f = await flows.validationPage.withAuthToken(token).start({});

  const runId = f.id;
  let stepId = f.steps[0].id;

  // Provide an action when no actions are defined on the page - this should fail the step
  f = await flows.validationPage
    .withAuthToken(token)
    .putStepValues(runId, stepId, { email: "test@example.com" }, "submit");

  // The step should have failed with an error
  expect(f.steps[0].status).toBe("FAILED");
  expect(f.steps[0].error).toContain(
    'invalid action "submit". No actions are defined for this page'
  );
});

test("flows - all inputs", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  const res = await flows.allInputs.withAuthToken(token).start({
    text: "text",
    number: 1,
    file: "data:text/plain;name=my-file.txt;base64,aGVsbG8=",
    date: new Date("2021-01-01"),
    timestamp: new Date("2021-01-01T12:30:15.000Z"),
    duration: new Duration("PT1000S"),
    bool: true,
    decimal: 1.1,
    myEnum: MyEnum.Value1,
    markdown: "**Hello**",
  });

  expect(res).toEqual({
    id: expect.any(String),
    status: "FAILED",
    name: "AllInputs",
    traceId: expect.any(String),
    startedBy: expect.any(String),
    input: {
      date: new Date("2021-01-01"),
      duration: "PT0S",
      file: "data:text/plain;name=my-file.txt;base64,aGVsbG8=",
      number: 1,
      text: "text",
      timestamp: new Date("2021-01-01T12:30:15.000Z"),
      bool: true,
      decimal: 1.1,
      myEnum: MyEnum.Value1,
      markdown: "**Hello**",
    },
    error: "date is not 2021-01-01",
    data: null,
    steps: [],
    config: {
      title: "All inputs",
    },
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
  });
});

test("flows - with completion", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  const res = await flows.withCompletion.withAuthToken(token).start({});
  expect(res).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "RUNNING",
    name: "WithCompletion",
    startedBy: expect.any(String),
    input: {},
    steps: [
      {
        id: expect.any(String),
        name: "my step",
        runId: res.id,
        stage: "starting",
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
    createdAt: res.createdAt,
    updatedAt: expect.any(Date),
    error: null,
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

  const flow = await flows.withCompletion
    .withAuthToken(token)
    .untilFinished(res.id);

  expect(flow).toEqual({
    id: res.id,
    traceId: res.traceId,
    status: "COMPLETED",
    name: "WithCompletion",
    startedBy: expect.any(String),
    error: null,
    input: {},
    steps: [
      {
        id: res.steps[0].id,
        name: "my step",
        runId: res.id,
        stage: "starting",
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
        name: "",
        runId: res.id,
        stage: "ending",
        status: "COMPLETED",
        type: "COMPLETE",
        value: null,
        error: null,
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
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
    createdAt: res.createdAt,
    updatedAt: expect.any(Date),
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
  const res = await flows.withCompletionMinimal.withAuthToken(token).start({});
  expect(res).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "COMPLETED",
    name: "WithCompletionMinimal",
    startedBy: expect.any(String),
    input: {},
    steps: [
      {
        id: expect.any(String),
        name: "",
        runId: res.id,
        stage: null,
        status: "COMPLETED",
        type: "COMPLETE",
        value: null,
        error: null,
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        ui: {
          __type: "ui.complete",
          title: "Completed flow",
          content: [],
        },
      },
    ],
    createdAt: res.createdAt,
    updatedAt: expect.any(Date),
    error: null,
    data: null,
    config: {
      title: "With completion minimal",
    },
  });

  const flow = await flows.withCompletionMinimal
    .withAuthToken(token)
    .get(res.id);

  expect(flow).toEqual({
    id: res.id,
    traceId: res.traceId,
    status: "COMPLETED",
    name: "WithCompletionMinimal",
    startedBy: expect.any(String),
    input: {},
    error: null,
    steps: [
      {
        id: expect.any(String),
        name: "",
        runId: res.id,
        stage: null,
        status: "COMPLETED",
        type: "COMPLETE",
        value: null,
        error: null,
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        ui: {
          __type: "ui.complete",
          title: "Completed flow",
          content: [],
        },
      },
    ],
    createdAt: res.createdAt,
    updatedAt: expect.any(Date),
    data: null,
    config: {
      title: "With completion minimal",
    },
  });
});

test("flows - with returned data", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  const res = await flows.withReturnedData.withAuthToken(token).start({});
  expect(res).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "RUNNING",
    name: "WithReturnedData",
    startedBy: expect.any(String),
    input: {},
    steps: [
      {
        id: expect.any(String),
        name: "my step",
        runId: res.id,
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
    createdAt: res.createdAt,
    updatedAt: expect.any(Date),
    error: null,
    data: null,
    config: {
      title: "With returned data",
    },
  });

  const flow = await flows.withReturnedData
    .withAuthToken(token)
    .untilFinished(res.id);

  expect(flow).toEqual({
    id: res.id,
    traceId: res.traceId,
    status: "COMPLETED",
    name: "WithReturnedData",
    startedBy: expect.any(String),
    input: {},
    steps: [
      {
        id: res.steps[0].id,
        name: "my step",
        runId: res.id,
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
    ],
    createdAt: res.createdAt,
    updatedAt: expect.any(Date),
    error: null,
    data: "hello",
    config: {
      title: "With returned data",
    },
  });
});

test("flows - myRuns", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  const res = await flows.errorInFlow.withAuthToken(token).start({});

  let f = await flows.scalarStep.withAuthToken(token).start({});

  await flows.scalarStep.withAuthToken(token).untilFinished(f.id);

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
  expect(resListAdmin.body.flows.length).toBe(23);
  expect(resListAdmin.body.flows[0].name).toBe("ScalarStep");
  expect(resListAdmin.body.flows[1].name).toBe("MixedStepTypes");
  expect(resListAdmin.body.flows[2].name).toBe("Stepless");
  expect(resListAdmin.body.flows[3].name).toBe("SingleStep");
  expect(resListAdmin.body.flows[4].name).toBe("ErrorInFlow");
  expect(resListAdmin.body.flows[5].name).toBe("OnlyPages");
  expect(resListAdmin.body.flows[6].name).toBe("OnlyFunctions");
  expect(resListAdmin.body.flows[7].name).toBe("ValidationText");
  expect(resListAdmin.body.flows[8].name).toBe("ValidationBoolean");
  expect(resListAdmin.body.flows[9].name).toBe("ValidationPage");
  expect(resListAdmin.body.flows[10].name).toBe("ValidationPageWithAction");
  expect(resListAdmin.body.flows[11].name).toBe("AllInputs");
  expect(resListAdmin.body.flows[12].name).toBe("EnvStep");
  expect(resListAdmin.body.flows[13].name).toBe("MultipleActions");
  expect(resListAdmin.body.flows[14].name).toBe("WithCompletion");
  expect(resListAdmin.body.flows[15].name).toBe("WithCompletionMinimal");
  expect(resListAdmin.body.flows[16].name).toBe("WithReturnedData");
  expect(resListAdmin.body.flows[17].name).toBe("ExpressionPermissionIsTrue");
  expect(resListAdmin.body.flows[18].name).toBe("DataWrapperConsistency");
  expect(resListAdmin.body.flows[19].name).toBe("MapStep");
  expect(resListAdmin.body.flows[20].name).toBe("DateStep");
  expect(resListAdmin.body.flows[21].name).toBe("ModelStep");
  expect(resListAdmin.body.flows[22].name).toBe("FileStep");

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
  await expect(
    flows.stepless.withAuthToken(token).start({})
  ).toHaveAuthorizationError();
});

test("flows - unauthenticated starting flow", async () => {
  await expect(flows.stepless.start({})).toHaveAuthorizationError();
});

test("flows - unauthorised getting flow", async () => {
  const adminToken = await getToken({ email: "admin@keel.xyz" });
  const resStart = await flows.stepless.withAuthToken(adminToken).start({});

  const userToken = await getToken({ email: "user@gmail.com" });
  await expect(
    flows.stepless.withAuthToken(userToken).get(resStart.id)
  ).toHaveAuthorizationError();
});

test("flows - unauthenticated getting flow", async () => {
  const adminToken = await getToken({ email: "admin@keel.xyz" });
  const resStart = await flows.stepless.withAuthToken(adminToken).start({});

  await expect(flows.stepless.get(resStart.id)).toHaveAuthorizationError();
});

test("flows - unauthenticated listing flows", async () => {
  const res = await listFlows({ token: null });
  expect(res.status).toBe(401);
});

test("flows - authorised starting flow with true expression", async () => {
  const token = await getToken({ email: "user@gmail.com" });
  const res = await flows.expressionPermissionIsTrue
    .withAuthToken(token)
    .start({});
  expect(res).not.toHaveAuthorizationError();
});

test("flows - not authorised starting flow with backlink expression", async () => {
  const token = await getToken({ email: "user@gmail.com" });
  await expect(
    flows.expressionPermissionCtx.withAuthToken(token).start({})
  ).toHaveAuthorizationError();
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

  await expect(
    flows.expressionPermissionCtx.withAuthToken(token).start({})
  ).toHaveAuthorizationError();
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

  const res = await flows.expressionPermissionCtx
    .withAuthToken(token)
    .start({});
  expect(res).not.toHaveAuthorizationError();
  expect(res.status).toBe("COMPLETED");
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

  const res = await flows.expressionPermissionEnv
    .withAuthToken(token)
    .start({});
  expect(res).not.toHaveAuthorizationError();
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

  await expect(
    flows.expressionPermissionEnv.withAuthToken(token).start({})
  ).toHaveAuthorizationError();
});

test("flows - unauthenticated starting flow with backlink expression", async () => {
  await expect(
    flows.expressionPermissionCtx.start({})
  ).toHaveAuthorizationError();
});

test("flows - env step", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  let f = await flows.envStep.withAuthToken(token).start({});

  const flow = await flows.envStep.withAuthToken(token).untilFinished(f.id);

  expect(flow).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "COMPLETED",
    name: "EnvStep",
    input: {},
    error: null,
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
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
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
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        ui: null,
      },
    ],
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "Env step",
    },
  });
});

test("flows - multiple actions - finish", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  let f = await flows.multipleActions.withAuthToken(token).start({});
  expect(f).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "AWAITING_INPUT",
    name: "MultipleActions",
    startedBy: expect.any(String),
    input: {},
    error: null,
    data: null,
    steps: [
      {
        id: expect.any(String),
        name: "question",
        runId: f.id,
        stage: null,
        status: "PENDING",
        type: "UI",
        value: null,
        error: null,
        startTime: expect.any(Date),
        endTime: null,
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
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
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "Multiple actions",
    },
  });

  f = await flows.multipleActions
    .withAuthToken(token)
    .putStepValues(f.id, f.steps[0].id, { yesno: true }, "finish");
  expect(f).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "COMPLETED",
    name: "MultipleActions",
    startedBy: expect.any(String),
    input: {},
    error: null,
    data: null,
    steps: expect.any(Array),
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "Multiple actions",
    },
  });
});

test("flows - multiple actions - continue", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  let f = await flows.multipleActions.withAuthToken(token).start({});
  expect(f).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "AWAITING_INPUT",
    name: "MultipleActions",
    startedBy: expect.any(String),
    input: {},
    error: null,
    data: null,
    steps: [
      {
        id: expect.any(String),
        name: "question",
        runId: f.id,
        stage: null,
        status: "PENDING",
        type: "UI",
        value: null,
        error: null,
        startTime: expect.any(Date),
        endTime: null,
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
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
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "Multiple actions",
    },
  });

  // Provide the values for the pending UI step
  f = await flows.multipleActions
    .withAuthToken(token)
    .putStepValues(f.id, f.steps[0].id, {}, "continue");

  expect(f).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "AWAITING_INPUT",
    name: "MultipleActions",
    startedBy: expect.any(String),
    input: {},
    error: null,
    data: null,
    steps: [
      {
        id: expect.any(String),
        name: "question",
        runId: f.id,
        stage: null,
        status: "COMPLETED",
        type: "UI",
        value: {},
        error: null,
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "another-question",
        runId: f.id,
        stage: null,
        status: "PENDING",
        type: "UI",
        value: null,
        error: null,
        startTime: expect.any(Date),
        endTime: null,
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
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
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "Multiple actions",
    },
  });

  f = await flows.multipleActions
    .withAuthToken(token)
    .putStepValues(f.id, f.steps[1].id, { name: "test" });

  expect(f).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "COMPLETED",
    name: "MultipleActions",
    startedBy: expect.any(String),
    input: {},
    error: null,
    data: null,
    steps: [
      {
        id: expect.any(String),
        name: "question",
        runId: f.id,
        stage: null,
        status: "COMPLETED",
        type: "UI",
        value: {},
        error: null,
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        ui: null,
      },
      {
        id: expect.any(String),
        name: "another-question",
        runId: f.id,
        stage: null,
        status: "COMPLETED",
        type: "UI",
        value: { name: "test" },
        error: null,
        startTime: expect.any(Date),
        endTime: expect.any(Date),
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        ui: null,
      },
    ],
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "Multiple actions",
    },
  });
});

test("flows - cancelling - with pending ui step", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  let f = await flows.mixedStepTypes.withAuthToken(token).start({
    name: "Keelson",
    age: 23,
  });

  const runId = f.id;
  const traceId = f.traceId;
  let step1 = f.steps[0];

  // The second step is a page with UI so we wait until the flow has reached that point
  f = await flows.mixedStepTypes.withAuthToken(token).untilAwaitingInput(runId);
  expect(f).toEqual({
    id: runId,
    name: "MixedStepTypes",
    startedBy: expect.any(String),
    status: "AWAITING_INPUT", // Flow is now awaiting input
    input: {
      name: "Keelson",
      age: 23,
    },
    error: null,
    data: null,
    config: {
      title: "Mixed step types",
    },
    traceId,
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
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
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: expect.any(Date),
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
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        startTime: expect.any(Date),
        endTime: null,
      },
    ],
  });

  step1 = f.steps[0];

  const resp = await flows.mixedStepTypes.withAuthToken(token).cancel(runId);
  expect(resp.status).toBe("CANCELLED"); // Flow run's status is CANCELLED
  expect(resp.steps[1]).toEqual({
    id: expect.any(String),
    runId: runId,
    stage: null,
    name: "confirm thing",
    error: null,
    ui: null, // We do not have the full UI config because this step is now cancelled
    status: "CANCELLED", // This step is now cancelled
    type: "UI",
    value: null,
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    startTime: expect.any(Date),
    endTime: expect.any(Date),
  });
});

test("flows - multiple actions - invalid action", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  let f = await flows.multipleActions.withAuthToken(token).start({});
  expect(f).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "AWAITING_INPUT",
    name: "MultipleActions",
    startedBy: expect.any(String),
    input: {},
    error: null,
    data: null,
    steps: [
      {
        id: expect.any(String),
        name: "question",
        runId: f.id,
        stage: null,
        status: "PENDING",
        type: "UI",
        value: null,
        error: null,
        startTime: expect.any(Date),
        endTime: null,
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
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
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "Multiple actions",
    },
  });

  // Provide the values for the pending UI step without an action - this should fail the step
  f = await flows.multipleActions
    .withAuthToken(token)
    .putStepValues(f.id, f.steps[0].id, {});

  // The step should have failed with an error
  expect(f.status).toBe("FAILED");
  expect(f.steps[0].status).toBe("FAILED");
  expect(f.steps[0].error).toContain(
    "action is required. Valid actions are: finish, continue"
  );
});

test("flows - stats", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  await flows.errorInFlow.withAuthToken(token).start({});

  let f = await flows.scalarStep.withAuthToken(token).start({});

  await flows.scalarStep.withAuthToken(token).untilFinished(f.id);

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
});

test("flows - data wrapper consistency", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  let f = await flows.dataWrapperConsistency.withAuthToken(token).start({});

  // First page: no actions - should be awaiting input
  expect(f).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "AWAITING_INPUT",
    name: "DataWrapperConsistency",
    startedBy: expect.any(String),
    input: {},
    error: null,
    data: null,
    steps: [
      {
        id: expect.any(String),
        name: "no-actions-page",
        runId: f.id,
        stage: null,
        status: "PENDING",
        type: "UI",
        value: null,
        error: null,
        startTime: expect.any(Date),
        endTime: null,
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        ui: expect.any(Object),
      },
    ],
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "Data wrapper consistency",
    },
  });

  // Submit first page (no actions) - data should NOT have wrapper
  f = await flows.dataWrapperConsistency
    .withAuthToken(token)
    .putStepValues(f.id, f.steps[0].id, { name: "John", age: 30 });

  expect(f.steps[0].status).toBe("COMPLETED");
  expect(f.steps[0].value).toEqual({ name: "John", age: 30 }); // No wrapper!
  expect(f.status).toBe("AWAITING_INPUT");

  const step1Id = f.steps[0].id;

  // Second page: with actions - should be awaiting input
  expect(f.steps[1].status).toBe("PENDING");
  expect(f.steps[1].ui).toHaveProperty("actions");

  // Submit second page (with actions) - should have wrapper when action is provided
  f = await flows.dataWrapperConsistency
    .withAuthToken(token)
    .putStepValues(f.id, f.steps[1].id, { city: "London" }, "next");

  expect(f.steps[1].status).toBe("COMPLETED");
  expect(f.steps[1].value).toEqual({ city: "London" }); // Just data, stored without wrapper

  // Now let's verify consistency by fetching the flow again
  const finalFlow = await flows.dataWrapperConsistency
    .withAuthToken(token)
    .untilFinished(f.id);

  // Check that when retrieved from DB, the data maintains the same structure
  expect(finalFlow.steps[0].value).toEqual({ name: "John", age: 30 }); // Still no wrapper
  expect(finalFlow.steps[1].value).toEqual({ city: "London" }); // Still no wrapper

  // Verify the final completion data has the expected structure
  expect(finalFlow.data).toHaveProperty("noActionsResult");
  expect(finalFlow.data).toHaveProperty("withActionsResult");

  // No actions result should be plain data
  expect(finalFlow.data.noActionsResult).toEqual({ name: "John", age: 30 });

  // With actions result should have { data, action } structure
  expect(finalFlow.data.withActionsResult).toHaveProperty("data");
  expect(finalFlow.data.withActionsResult).toHaveProperty("action");
  expect(finalFlow.data.withActionsResult.data).toEqual({ city: "London" });
  expect(finalFlow.data.withActionsResult.action).toBe("next");
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
