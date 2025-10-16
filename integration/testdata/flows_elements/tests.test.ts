import { resetDatabase, models, flows } from "@teamkeel/testing";
import { beforeEach, expect, test } from "vitest";

beforeEach(resetDatabase);
test("flows - callback flow", async () => {
  let flow = await flows.callbackFlow.start({});
  expect(flow).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "AWAITING_INPUT",
    name: "CallbackFlow",
    startedBy: null,
    input: {},
    error: null,
    data: null,
    steps: [
      {
        id: expect.any(String),
        name: "my page",
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
            {
              __type: "ui.input.number",
              defaultValue: 1,
              disabled: false,
              label: "How many numbers?",
              name: "numberInput",
              optional: false,
            },
            {
              __type: "ui.input.boolean",
              disabled: false,
              label: "True?",
              mode: "checkbox",
              name: "boolInput",
              optional: false,
            },
          ],
          hasValidationErrors: false,
        },
      },
    ],
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "Callback flow",
    },
  });

  let callbackResponse = await flows.callbackFlow.callback(
    flow.id,
    flow.steps[0].id,
    "numberInput",
    "onLeave",
    12
  );
  expect(callbackResponse).toEqual(24);

  callbackResponse = await flows.callbackFlow.callback(
    flow.id,
    flow.steps[0].id,
    "numberInput",
    "onLeave",
    50
  );
  expect(callbackResponse).toEqual(100);

  callbackResponse = await flows.callbackFlow.callback(
    flow.id,
    flow.steps[0].id,
    "boolInput",
    "onLeave",
    false
  );
  expect(callbackResponse).toEqual(true);

  await expect(
    flows.callbackFlow.callback(
      flow.id,
      flow.steps[0].id,
      "wrong",
      "onLeave",
      false
    )
  ).toHaveError({
    code: "ERR_UNKNOWN",
    message: "Element with name wrong not found",
  });
});

test("flows - bulkScan element", async () => {
  let flow = await flows.bulkScan.start({});
  expect(flow).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "AWAITING_INPUT",
    name: "BulkScan",
    startedBy: null,
    input: {},
    error: null,
    data: null,
    steps: [
      {
        id: expect.any(String),
        name: "multi scan page",
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
            {
              __type: "ui.input.scan",
              duplicateHandling: "rejectDuplicates",
              mode: "multi",
              name: "bulkScan",
            },
          ],
          hasValidationErrors: false,
        },
      },
    ],
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "Bulk scan",
    },
  });

  // Provide the values for the pending UI step
  flow = await flows.bulkScan.putStepValues(flow.id, flow.steps[0].id, {
    bulkScan: ["123", "456", "789"],
  });
  expect(flow).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "AWAITING_INPUT",
    name: "BulkScan",
    startedBy: null,
    input: {},
    error: null,
    data: null,
    steps: [
      {
        id: expect.any(String),
        name: "multi scan page",
        runId: expect.any(String),
        stage: null,
        status: "COMPLETED",
        type: "UI",
        value: {
          bulkScan: ["123", "456", "789"],
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
        name: "single scan page",
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
          actions: [
            {
              label: "finish",
              mode: "primary",
              value: "finish",
            },
          ],
          content: [
            {
              __type: "ui.input.scan",
              duplicateHandling: "none",
              mode: "single",
              name: "singleScan",
            },
          ],
          hasValidationErrors: false,
        },
      },
    ],
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "Bulk scan",
    },
  });

  // Provide the values for the pending UI step
  flow = await flows.bulkScan.putStepValues(
    flow.id,
    flow.steps[1].id,
    {
      singleScan: "abc",
    },
    "finish"
  );
  expect(flow).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "COMPLETED",
    name: "BulkScan",
    startedBy: null,
    input: {},
    error: null,
    data: null,
    steps: [
      {
        id: expect.any(String),
        name: "multi scan page",
        runId: expect.any(String),
        stage: null,
        status: "COMPLETED",
        type: "UI",
        value: {
          bulkScan: ["123", "456", "789"],
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
        name: "single scan page",
        runId: expect.any(String),
        stage: null,
        status: "COMPLETED",
        type: "UI",
        value: {
          singleScan: "abc",
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
      title: "Bulk scan",
    },
  });
});

test("flows - iterator element", async () => {
  let flow = await flows.iterator.start({});

  expect(flow).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "AWAITING_INPUT",
    name: "Iterator",
    startedBy: null,
    input: {},
    error: null,
    data: null,
    steps: [
      {
        id: expect.any(String),
        name: "my page",
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
            {
              __type: "ui.iterator",
              content: [
                {
                  __type: "ui.display.header",
                  description: "my description",
                  level: 1,
                  title: "my header",
                },
                {
                  __type: "ui.select.one",
                  disabled: false,
                  label: "SKU",
                  name: "sku",
                  optional: false,
                  options: [
                    "SHOES",
                    "SHIRTS",
                    "PANTS",
                    "TIE",
                    "BELT",
                    "SOCKS",
                    "UNDERWEAR",
                  ],
                },
                {
                  __type: "ui.input.number",
                  disabled: false,
                  label: "Qty",
                  name: "quantity",
                  optional: false,
                },
              ],
              min: 1,
              name: "my iterator",
            },
          ],
          hasValidationErrors: false,
        },
      },
    ],
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "Iterator",
    },
  });

  // Provide the values for the pending UI step
  flow = await flows.iterator.putStepValues(flow.id, flow.steps[0].id, {
    "my iterator": [
      {
        sku: "SHOES",
        quantity: 1,
      },
      {
        sku: "SHIRTS",
        quantity: 5,
      },
      {
        sku: "PANTS",
        quantity: 3,
      },
    ],
  });
  expect(flow).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "COMPLETED",
    name: "Iterator",
    startedBy: null,
    input: {},
    error: null,
    data: null,
    steps: [
      {
        id: expect.any(String),
        name: "my page",
        runId: expect.any(String),
        stage: null,
        status: "COMPLETED",
        type: "UI",
        value: {
          "my iterator": [
            {
              sku: "SHOES",
              quantity: 1,
            },
            {
              sku: "SHIRTS",
              quantity: 5,
            },
            {
              sku: "PANTS",
              quantity: 3,
            },
          ],
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
      title: "Iterator",
    },
  });
});

test("flows - iterator element - iterator and element validation errors", async () => {
  let flow = await flows.iterator.start({});

  flow = await flows.iterator.putStepValues(flow.id, flow.steps[0].id, {
    "my iterator": [
      {
        sku: "SHOES",
        quantity: 1,
      },
      {
        sku: "SHIRTS",
        quantity: 0,
      },
      {
        sku: "SHIRTS",
        quantity: 30,
      },
    ],
  });

  expect(flow).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "AWAITING_INPUT",
    name: "Iterator",
    startedBy: null,
    input: {},
    error: null,
    data: null,
    steps: [
      {
        id: expect.any(String),
        name: "my page",
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
            {
              __type: "ui.iterator",
              content: [
                {
                  __type: "ui.display.header",
                  description: "my description",
                  level: 1,
                  title: "my header",
                },
                {
                  __type: "ui.select.one",
                  disabled: false,
                  label: "SKU",
                  name: "sku",
                  optional: false,
                  options: [
                    "SHOES",
                    "SHIRTS",
                    "PANTS",
                    "TIE",
                    "BELT",
                    "SOCKS",
                    "UNDERWEAR",
                  ],
                },
                {
                  __type: "ui.input.number",
                  disabled: false,
                  label: "Qty",
                  name: "quantity",
                  optional: false,
                },
              ],
              min: 1,
              name: "my iterator",
              validationError: "SHIRTS has been selected twice",
              contentValidationErrors: [
                {
                  index: 1,
                  name: "quantity",
                  validationError: "Quantity must be greater than 0",
                },
                {
                  index: 2,
                  name: "quantity",
                  validationError: "Quantity must be less than 10",
                },
              ],
            },
          ],
          hasValidationErrors: true,
          validationError: "Total quantity must be less than 20",
        },
      },
    ],
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "Iterator",
    },
  });
});

test("flows - pickList element with validation", async () => {
  let flow = await flows.pickListValidation.start({});

  expect(flow).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "AWAITING_INPUT",
    name: "PickListValidation",
    startedBy: null,
    input: {},
    error: null,
    data: null,
    steps: [
      {
        id: expect.any(String),
        name: "pick list page",
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
            {
              __type: "ui.interactive.pickList",
              name: "items",
              data: [
                {
                  id: "prod-1",
                  targetQuantity: 10,
                  title: "Widget A",
                  barcodes: ["1234567890"],
                },
                {
                  id: "prod-2",
                  targetQuantity: 5,
                  title: "Widget B",
                  barcodes: ["0987654321"],
                },
                {
                  id: "prod-3",
                  targetQuantity: 3,
                  title: "Widget C",
                  barcodes: ["1111111111"],
                },
              ],
              supportedInputs: {
                scanner: true,
                manual: true,
              },
            },
          ],
          hasValidationErrors: false,
        },
      },
    ],
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "Pick list validation",
    },
  });

  // Test validation error: total quantity exceeds limit
  flow = await flows.pickListValidation.putStepValues(
    flow.id,
    flow.steps[0].id,
    {
      items: {
        items: [
          { id: "prod-1", quantity: 10, targetQuantity: 10 },
          { id: "prod-2", quantity: 8, targetQuantity: 5 },
          { id: "prod-3", quantity: 3, targetQuantity: 3 },
        ],
      },
    }
  );

  expect(flow.steps[0].status).toBe("PENDING");
  expect(flow.steps[0].ui).toMatchObject({
    __type: "ui.page",
    hasValidationErrors: true,
    content: [
      {
        __type: "ui.interactive.pickList",
        name: "items",
        validationError: "Total quantity cannot exceed 20 items",
      },
    ],
  });

  // Test validation error: no items picked
  flow = await flows.pickListValidation.putStepValues(
    flow.id,
    flow.steps[0].id,
    {
      items: {
        items: [
          { id: "prod-1", quantity: 0, targetQuantity: 10 },
          { id: "prod-2", quantity: 0, targetQuantity: 5 },
          { id: "prod-3", quantity: 0, targetQuantity: 3 },
        ],
      },
    }
  );

  expect(flow.steps[0].status).toBe("PENDING");
  expect(flow.steps[0].ui).toMatchObject({
    __type: "ui.page",
    hasValidationErrors: true,
    content: [
      {
        __type: "ui.interactive.pickList",
        name: "items",
        validationError: "At least one item must be picked",
      },
    ],
  });

  // Test successful validation
  flow = await flows.pickListValidation.putStepValues(
    flow.id,
    flow.steps[0].id,
    {
      items: {
        items: [
          { id: "prod-1", quantity: 8, targetQuantity: 10 },
          { id: "prod-2", quantity: 5, targetQuantity: 5 },
          { id: "prod-3", quantity: 2, targetQuantity: 3 },
        ],
      },
    }
  );

  expect(flow).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "COMPLETED",
    name: "PickListValidation",
    startedBy: null,
    input: {},
    error: null,
    data: {
      items: [
        { id: "prod-1", quantity: 8, targetQuantity: 10 },
        { id: "prod-2", quantity: 5, targetQuantity: 5 },
        { id: "prod-3", quantity: 2, targetQuantity: 3 },
      ],
    },
    steps: [
      {
        id: expect.any(String),
        name: "pick list page",
        runId: expect.any(String),
        stage: null,
        status: "COMPLETED",
        type: "UI",
        value: {
          items: {
            items: [
              { id: "prod-1", quantity: 8, targetQuantity: 10 },
              { id: "prod-2", quantity: 5, targetQuantity: 5 },
              { id: "prod-3", quantity: 2, targetQuantity: 3 },
            ],
          },
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
      title: "Pick list validation",
    },
  });
});
