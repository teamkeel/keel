import { resetDatabase, models, flows } from "@teamkeel/testing";
import { beforeEach, expect, test } from "vitest";

beforeEach(resetDatabase);

test("flows - iterator element", async () => {
  let flow = await flows.iterator.start({});

  expect(flow).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "AWAITING_INPUT",
    name: "Iterator",
    startedBy: null,
    input: {},
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
  flow = await flows.iterator.putStepValues(
    flow.id,
    flow.steps[0].id,

    {
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
    }
  );
  expect(flow).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "COMPLETED",
    name: "Iterator",
    startedBy: null,
    input: {},
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
