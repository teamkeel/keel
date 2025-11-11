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
              autoContinue: false,
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
              autoContinue: false,
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
        autoContinue: false,
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
        autoContinue: false,
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

test("flows - dataGrid element - basic with inferred columns", async () => {
  let flow = await flows.dataGridValidation.start({});

  expect(flow).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "AWAITING_INPUT",
    name: "DataGridValidation",
    startedBy: null,
    input: {},
    error: null,
    data: null,
    steps: [
      {
        id: expect.any(String),
        name: "basic data grid",
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
              __type: "ui.input.dataGrid",
              name: "products",
              data: [
                { id: "prod-1", name: "Widget A", quantity: 10, inStock: true },
                { id: "prod-2", name: "Widget B", quantity: 5, inStock: false },
                { id: "prod-3", name: "Widget C", quantity: 0, inStock: true },
              ],
              columns: [
                {
                  key: "id",
                  label: "Id",
                  index: 0,
                  type: "text",
                  editable: true,
                },
                {
                  key: "name",
                  label: "Name",
                  index: 1,
                  type: "text",
                  editable: true,
                },
                {
                  key: "quantity",
                  label: "Quantity",
                  index: 2,
                  type: "number",
                  editable: true,
                },
                {
                  key: "inStock",
                  label: "In stock",
                  index: 3,
                  type: "boolean",
                  editable: true,
                },
              ],
              allowAddRows: false,
              allowDeleteRows: false,
            },
          ],
          hasValidationErrors: false,
        },
      },
    ],
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "Data grid validation",
    },
  });

  // Submit valid data
  flow = await flows.dataGridValidation.putStepValues(
    flow.id,
    flow.steps[0].id,
    {
      products: [
        { id: "prod-1", name: "Widget A", quantity: 15, inStock: true },
        { id: "prod-2", name: "Widget B", quantity: 8, inStock: true },
      ],
    }
  );

  expect(flow.steps[0].status).toBe("COMPLETED");
  expect(flow.steps[1].status).toBe("PENDING");
  expect(flow.steps[1].name).toBe("data grid with validation");
});

test("flows - dataGrid element - with explicit columns and validation", async () => {
  let flow = await flows.dataGridValidation.start({});

  // Skip first page
  flow = await flows.dataGridValidation.putStepValues(
    flow.id,
    flow.steps[0].id,
    {
      products: [
        { id: "prod-1", name: "Widget A", quantity: 10, inStock: true },
      ],
    }
  );

  expect(flow.steps[1].ui).toMatchObject({
    __type: "ui.page",
    content: [
      {
        __type: "ui.input.dataGrid",
        name: "inventory",
        data: [
          { id: "item-1", sku: "SKU001", quantity: 10, price: 99.99 },
          { id: "item-2", sku: "SKU002", quantity: 5, price: 149.99 },
          { id: "item-3", sku: "SKU003", quantity: 0, price: 199.99 },
        ],
        columns: [
          { key: "id", label: "ID", index: 0, type: "id", editable: false },
          { key: "sku", label: "SKU", index: 1, type: "text", editable: true },
          {
            key: "quantity",
            label: "Qty",
            index: 2,
            type: "number",
            editable: true,
          },
          {
            key: "price",
            label: "Price",
            index: 3,
            type: "number",
            editable: true,
          },
        ],
        allowAddRows: true,
        allowDeleteRows: true,
      },
    ],
    hasValidationErrors: false,
  });

  // Test validation error: total quantity exceeds limit
  flow = await flows.dataGridValidation.putStepValues(
    flow.id,
    flow.steps[1].id,
    {
      inventory: [
        { id: "item-1", sku: "SKU001", quantity: 50, price: 99.99 },
        { id: "item-2", sku: "SKU002", quantity: 55, price: 149.99 },
      ],
    }
  );

  expect(flow.steps[1].status).toBe("PENDING");
  expect(flow.steps[1].ui).toMatchObject({
    __type: "ui.page",
    hasValidationErrors: true,
    content: [
      {
        __type: "ui.input.dataGrid",
        name: "inventory",
        validationError: "Total quantity cannot exceed 100 items",
      },
    ],
  });

  // Test validation error: negative quantity
  flow = await flows.dataGridValidation.putStepValues(
    flow.id,
    flow.steps[1].id,
    {
      inventory: [
        { id: "item-1", sku: "SKU001", quantity: -5, price: 99.99 },
        { id: "item-2", sku: "SKU002", quantity: 10, price: 149.99 },
      ],
    }
  );

  expect(flow.steps[1].status).toBe("PENDING");
  expect(flow.steps[1].ui).toMatchObject({
    __type: "ui.page",
    hasValidationErrors: true,
    content: [
      {
        __type: "ui.input.dataGrid",
        name: "inventory",
        validationError: "Quantities must be non-negative",
      },
    ],
  });

  // Test validation error: empty array
  flow = await flows.dataGridValidation.putStepValues(
    flow.id,
    flow.steps[1].id,
    {
      inventory: [],
    }
  );

  expect(flow.steps[1].status).toBe("PENDING");
  expect(flow.steps[1].ui).toMatchObject({
    __type: "ui.page",
    hasValidationErrors: true,
    content: [
      {
        __type: "ui.input.dataGrid",
        name: "inventory",
        validationError: "At least one item must be present",
      },
    ],
  });

  // Test validation error: invalid price
  flow = await flows.dataGridValidation.putStepValues(
    flow.id,
    flow.steps[1].id,
    {
      inventory: [
        { id: "item-1", sku: "SKU001", quantity: 10, price: 0 },
        { id: "item-2", sku: "SKU002", quantity: 5, price: 149.99 },
      ],
    }
  );

  expect(flow.steps[1].status).toBe("PENDING");
  expect(flow.steps[1].ui).toMatchObject({
    __type: "ui.page",
    hasValidationErrors: true,
    content: [
      {
        __type: "ui.input.dataGrid",
        name: "inventory",
        validationError: "All prices must be greater than zero",
      },
    ],
  });

  // Test successful validation
  flow = await flows.dataGridValidation.putStepValues(
    flow.id,
    flow.steps[1].id,
    {
      inventory: [
        { id: "item-1", sku: "SKU001", quantity: 25, price: 99.99 },
        { id: "item-2", sku: "SKU002", quantity: 30, price: 149.99 },
        { id: "item-3", sku: "SKU003", quantity: 20, price: 199.99 },
      ],
    }
  );

  expect(flow.steps[1].status).toBe("COMPLETED");
  expect(flow.steps[2].status).toBe("PENDING");
  expect(flow.steps[2].name).toBe("data grid with types");
});

test("flows - dataGrid element - type coercion", async () => {
  let flow = await flows.dataGridValidation.start({});

  // Skip first two pages
  flow = await flows.dataGridValidation.putStepValues(
    flow.id,
    flow.steps[0].id,
    {
      products: [
        { id: "prod-1", name: "Widget A", quantity: 10, inStock: true },
      ],
    }
  );

  flow = await flows.dataGridValidation.putStepValues(
    flow.id,
    flow.steps[1].id,
    {
      inventory: [{ id: "item-1", sku: "SKU001", quantity: 10, price: 99.99 }],
    }
  );

  // Check third page structure
  expect(flow.steps[2].ui).toMatchObject({
    __type: "ui.page",
    content: [
      {
        __type: "ui.input.dataGrid",
        name: "orders",
        data: [
          {
            orderId: "ORD-001",
            customerName: "John Doe",
            orderTotal: 250.5,
            isPaid: true,
          },
          {
            orderId: "ORD-002",
            customerName: "Jane Smith",
            orderTotal: 125.75,
            isPaid: false,
          },
        ],
        columns: [
          {
            key: "orderId",
            label: "Order ID",
            index: 0,
            type: "text",
            editable: true,
          },
          {
            key: "customerName",
            label: "Customer",
            index: 1,
            type: "text",
            editable: true,
          },
          {
            key: "orderTotal",
            label: "Total",
            index: 2,
            type: "number",
            editable: true,
          },
          {
            key: "isPaid",
            label: "Paid",
            index: 3,
            type: "boolean",
            editable: true,
          },
        ],
        allowAddRows: false,
        allowDeleteRows: false,
      },
    ],
    hasValidationErrors: false,
  });

  // Submit valid data
  flow = await flows.dataGridValidation.putStepValues(
    flow.id,
    flow.steps[2].id,
    {
      orders: [
        {
          orderId: "ORD-001",
          customerName: "John Doe",
          orderTotal: 300,
          isPaid: true,
        },
      ],
    }
  );

  expect(flow).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "COMPLETED",
    name: "DataGridValidation",
    startedBy: null,
    input: {},
    error: null,
    data: null,
    steps: [
      {
        id: expect.any(String),
        name: "basic data grid",
        runId: expect.any(String),
        stage: null,
        status: "COMPLETED",
        type: "UI",
        value: {
          products: [
            { id: "prod-1", name: "Widget A", quantity: 10, inStock: true },
          ],
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
        name: "data grid with validation",
        runId: expect.any(String),
        stage: null,
        status: "COMPLETED",
        type: "UI",
        value: {
          inventory: [
            { id: "item-1", sku: "SKU001", quantity: 10, price: 99.99 },
          ],
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
        name: "data grid with types",
        runId: expect.any(String),
        stage: null,
        status: "COMPLETED",
        type: "UI",
        value: {
          orders: [
            {
              orderId: "ORD-001",
              customerName: "John Doe",
              orderTotal: 300,
              isPaid: true,
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
      title: "Data grid validation",
    },
  });
});

// ============================================================================
// BASIC ELEMENT TESTS - Elements Without Actions
// ============================================================================

test("flows - text input element - validation", async () => {
  let flow = await flows.textInput.start({});

  expect(flow.status).toBe("AWAITING_INPUT");
  expect(flow.steps[0].status).toBe("PENDING");

  // Test username validation: too short
  flow = await flows.textInput.putStepValues(flow.id, flow.steps[0].id, {
    username: "ab",
    email: "test@example.com",
    description: "Valid description here",
  });

  expect(flow.steps[0].status).toBe("PENDING");
  expect((flow.steps[0].ui as any)?.hasValidationErrors).toBe(true);
  expect((flow.steps[0].ui as any)?.content[0].validationError).toEqual(
    "Username must be at least 3 characters"
  );

  // Test username validation: contains spaces
  flow = await flows.textInput.putStepValues(flow.id, flow.steps[0].id, {
    username: "john doe",
    email: "test@example.com",
    description: "Valid description here",
  });

  expect(flow.steps[0].status).toBe("PENDING");
  expect((flow.steps[0].ui as any)?.hasValidationErrors).toBe(true);
  expect((flow.steps[0].ui as any)?.content[0].validationError).toEqual(
    "Username cannot contain spaces"
  );

  // Test email validation: missing @
  flow = await flows.textInput.putStepValues(flow.id, flow.steps[0].id, {
    username: "john_doe",
    email: "invalidemail",
    description: "Valid description here",
  });

  expect(flow.steps[0].status).toBe("PENDING");
  expect((flow.steps[0].ui as any)?.hasValidationErrors).toBe(true);
  expect((flow.steps[0].ui as any)?.content[1].validationError).toEqual(
    "Email must contain @"
  );

  // Test description validation: too short
  flow = await flows.textInput.putStepValues(flow.id, flow.steps[0].id, {
    username: "john_doe",
    email: "john@example.com",
    description: "Hi",
  });

  expect(flow.steps[0].status).toBe("PENDING");
  expect((flow.steps[0].ui as any)?.hasValidationErrors).toBe(true);
  expect((flow.steps[0].ui as any)?.content[2].validationError).toEqual(
    "Description must be at least 5 characters"
  );

  // Test description validation: too long
  flow = await flows.textInput.putStepValues(flow.id, flow.steps[0].id, {
    username: "john_doe",
    email: "john@example.com",
    description: "a".repeat(101),
  });

  expect(flow.steps[0].status).toBe("PENDING");
  expect((flow.steps[0].ui as any)?.hasValidationErrors).toBe(true);
  expect((flow.steps[0].ui as any)?.content[2].validationError).toEqual(
    "Description must be at most 100 characters"
  );

  // Test successful validation with all valid values
  flow = await flows.textInput.putStepValues(flow.id, flow.steps[0].id, {
    username: "john_doe",
    email: "john@example.com",
    description: "Valid test user description",
  });

  expect(flow.status).toBe("COMPLETED");
  expect(flow.steps[0].status).toBe("COMPLETED");
  expect(flow.steps[0].value).toEqual({
    username: "john_doe",
    email: "john@example.com",
    description: "Valid test user description",
  });
});

test("flows - number input element - validation", async () => {
  let flow = await flows.numberInput.start({});

  expect(flow.status).toBe("AWAITING_INPUT");
  expect(flow.steps[0].status).toBe("PENDING");

  // Test age validation: negative
  flow = await flows.numberInput.putStepValues(flow.id, flow.steps[0].id, {
    age: -5,
    quantity: 10,
    price: 50,
  });

  expect(flow.steps[0].status).toBe("PENDING");
  expect((flow.steps[0].ui as any)?.hasValidationErrors).toBe(true);
  expect((flow.steps[0].ui as any)?.content[0].validationError).toEqual(
    "Age cannot be negative"
  );

  // Test age validation: too high
  flow = await flows.numberInput.putStepValues(flow.id, flow.steps[0].id, {
    age: 200,
    quantity: 10,
    price: 50,
  });

  expect(flow.steps[0].status).toBe("PENDING");
  expect((flow.steps[0].ui as any)?.hasValidationErrors).toBe(true);
  expect((flow.steps[0].ui as any)?.content[0].validationError).toEqual(
    "Age must be 150 or less"
  );

  // Test quantity validation: less than 1
  flow = await flows.numberInput.putStepValues(flow.id, flow.steps[0].id, {
    age: 30,
    quantity: 0,
    price: 50,
  });

  expect(flow.steps[0].status).toBe("PENDING");
  expect((flow.steps[0].ui as any)?.hasValidationErrors).toBe(true);
  expect((flow.steps[0].ui as any)?.content[1].validationError).toEqual(
    "Quantity must be at least 1"
  );

  // Test price validation: negative
  flow = await flows.numberInput.putStepValues(flow.id, flow.steps[0].id, {
    age: 30,
    quantity: 5,
    price: -10,
  });

  expect(flow.steps[0].status).toBe("PENDING");
  expect((flow.steps[0].ui as any)?.hasValidationErrors).toBe(true);
  expect((flow.steps[0].ui as any)?.content[2].validationError).toEqual(
    "Price cannot be negative"
  );

  // Test successful validation with all valid values
  flow = await flows.numberInput.putStepValues(flow.id, flow.steps[0].id, {
    age: 30,
    quantity: 5,
    price: 99.99,
  });

  expect(flow.status).toBe("COMPLETED");
  expect(flow.steps[0].status).toBe("COMPLETED");
  expect(flow.steps[0].value).toEqual({
    age: 30,
    quantity: 5,
    price: 99.99,
  });
});

test("flows - boolean input element - validation", async () => {
  let flow = await flows.booleanInput.start({});

  expect(flow.status).toBe("AWAITING_INPUT");
  expect(flow.steps[0].status).toBe("PENDING");

  // Test isActive validation: must be true
  flow = await flows.booleanInput.putStepValues(flow.id, flow.steps[0].id, {
    isActive: false,
    agreedToTerms: true,
    receiveNewsletter: false,
  });

  expect(flow.steps[0].status).toBe("PENDING");
  expect((flow.steps[0].ui as any)?.hasValidationErrors).toBe(true);
  expect((flow.steps[0].ui as any)?.content[0].validationError).toEqual(
    "Account must be active"
  );

  // Test agreedToTerms validation: if provided, must be true
  flow = await flows.booleanInput.putStepValues(flow.id, flow.steps[0].id, {
    isActive: true,
    agreedToTerms: false,
    receiveNewsletter: false,
  });

  expect(flow.steps[0].status).toBe("PENDING");
  expect((flow.steps[0].ui as any)?.hasValidationErrors).toBe(true);
  expect((flow.steps[0].ui as any)?.content[1].validationError).toEqual(
    "You must agree to the terms"
  );

  // Test successful validation with all valid values
  flow = await flows.booleanInput.putStepValues(flow.id, flow.steps[0].id, {
    isActive: true,
    agreedToTerms: true,
    receiveNewsletter: false,
  });

  expect(flow.status).toBe("COMPLETED");
  expect(flow.steps[0].status).toBe("COMPLETED");
  expect(flow.steps[0].value).toEqual({
    isActive: true,
    agreedToTerms: true,
    receiveNewsletter: false,
  });
});

test("flows - date picker element - validation", async () => {
  let flow = await flows.datePickerInput.start({});

  expect(flow.status).toBe("AWAITING_INPUT");
  expect(flow.steps[0].status).toBe("PENDING");

  // Test birthDate validation: future date
  const tomorrow = new Date();
  tomorrow.setDate(tomorrow.getDate() + 1);

  flow = await flows.datePickerInput.putStepValues(flow.id, flow.steps[0].id, {
    birthDate: tomorrow,
    startDate: new Date("2025-06-15"),
    appointmentDate: new Date("2025-12-01"),
  });

  expect(flow.steps[0].status).toBe("PENDING");
  expect((flow.steps[0].ui as any)?.hasValidationErrors).toBe(true);
  expect((flow.steps[0].ui as any)?.content[0].validationError).toEqual(
    "Birth date must be in the past"
  );

  // Test birthDate validation: too old
  flow = await flows.datePickerInput.putStepValues(flow.id, flow.steps[0].id, {
    birthDate: new Date("1800-01-01"),
    startDate: new Date("2025-06-15"),
    appointmentDate: new Date("2025-12-01"),
  });

  expect(flow.steps[0].status).toBe("PENDING");
  expect((flow.steps[0].ui as any)?.hasValidationErrors).toBe(true);
  expect((flow.steps[0].ui as any)?.content[0].validationError).toEqual(
    "Birth date must be after 1900"
  );

  // Test startDate validation: more than 1 year in future
  const twoYearsFromNow = new Date();
  twoYearsFromNow.setFullYear(twoYearsFromNow.getFullYear() + 2);

  flow = await flows.datePickerInput.putStepValues(flow.id, flow.steps[0].id, {
    birthDate: new Date("1990-05-20"),
    startDate: twoYearsFromNow,
    appointmentDate: new Date("2025-12-01"),
  });

  expect(flow.steps[0].status).toBe("PENDING");
  expect((flow.steps[0].ui as any)?.hasValidationErrors).toBe(true);
  expect((flow.steps[0].ui as any)?.content[1].validationError).toEqual(
    "Start date cannot be more than 1 year in the future"
  );

  // Test appointmentDate validation: past date
  const yesterday = new Date();
  yesterday.setDate(yesterday.getDate() - 1);

  flow = await flows.datePickerInput.putStepValues(flow.id, flow.steps[0].id, {
    birthDate: new Date("1990-05-20"),
    startDate: new Date("2025-06-15"),
    appointmentDate: yesterday,
  });

  expect(flow.steps[0].status).toBe("PENDING");
  expect((flow.steps[0].ui as any)?.hasValidationErrors).toBe(true);
  expect((flow.steps[0].ui as any)?.content[2].validationError).toEqual(
    "Appointment date must be in the future"
  );

  // Test successful validation with all valid values
  const futureDate = new Date();
  futureDate.setDate(futureDate.getDate() + 30);

  flow = await flows.datePickerInput.putStepValues(flow.id, flow.steps[0].id, {
    birthDate: new Date("1990-05-20"),
    startDate: new Date("2025-06-15"),
    appointmentDate: futureDate,
  });

  expect(flow.status).toBe("COMPLETED");
  expect(flow.steps[0].status).toBe("COMPLETED");
  expect(flow.steps[0].value).toEqual({
    birthDate: new Date("1990-05-20"),
    startDate: new Date("2025-06-15"),
    appointmentDate: futureDate,
  });
});

test("flows - single scan element - validation", async () => {
  let flow = await flows.singleScan.start({});

  expect(flow.status).toBe("AWAITING_INPUT");
  expect(flow.steps[0].status).toBe("PENDING");

  // Test barcode validation: too short
  flow = await flows.singleScan.putStepValues(flow.id, flow.steps[0].id, {
    barcode: "1234567",
  });

  expect(flow.steps[0].status).toBe("PENDING");
  expect((flow.steps[0].ui as any)?.hasValidationErrors).toBe(true);
  expect((flow.steps[0].ui as any)?.content[0].validationError).toEqual(
    "Barcode must be at least 8 characters"
  );

  // Test barcode validation: not alphanumeric
  flow = await flows.singleScan.putStepValues(flow.id, flow.steps[0].id, {
    barcode: "12345678-ABC",
  });

  expect(flow.steps[0].status).toBe("PENDING");
  expect((flow.steps[0].ui as any)?.hasValidationErrors).toBe(true);
  expect((flow.steps[0].ui as any)?.content[0].validationError).toEqual(
    "Barcode must be alphanumeric"
  );

  // Test successful validation with valid barcode
  flow = await flows.singleScan.putStepValues(flow.id, flow.steps[0].id, {
    barcode: "ABC12345678",
  });

  expect(flow.status).toBe("COMPLETED");
  expect(flow.steps[0].status).toBe("COMPLETED");
  expect(flow.steps[0].value).toEqual({
    barcode: "ABC12345678",
  });
});

// ============================================================================
// VALIDATION WITH ACTIONS TESTS - Individual Element Tests
// ============================================================================

test("flows - page validation with actions - submit requires valid data", async () => {
  let flow = await flows.pageValidationWithActions.start({});

  // Submit action should fail when name is missing
  flow = await flows.pageValidationWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    { name: "", age: 20 },
    "submit"
  );

  expect((flow.steps[0].ui as any)?.validationError).toEqual(
    "Name is required when submitting"
  );
  expect((flow.steps[0].ui as any)?.hasValidationErrors).toBe(true);
  expect(flow.steps[0].status).toBe("PENDING");

  // Submit action should fail when age < 18
  flow = await flows.pageValidationWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    { name: "John", age: 15 },
    "submit"
  );

  expect((flow.steps[0].ui as any)?.validationError).toEqual(
    "Must be 18 or older to submit"
  );
  expect((flow.steps[0].ui as any)?.hasValidationErrors).toBe(true);
  expect(flow.steps[0].status).toBe("PENDING");

  // Submit action should pass with valid data
  flow = await flows.pageValidationWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    { name: "John", age: 25 },
    "submit"
  );

  expect(flow.steps[0].status).toBe("COMPLETED");
  expect(flow.status).toBe("COMPLETED");
});

test("flows - page validation with actions - draft allows any data", async () => {
  let flow = await flows.pageValidationWithActions.start({});

  // Draft action should pass even with invalid data
  flow = await flows.pageValidationWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    { name: "", age: 10 },
    "draft"
  );

  expect(flow.steps[0].status).toBe("COMPLETED");
  expect(flow.status).toBe("COMPLETED");
});

test("flows - text input with actions - submit validates format", async () => {
  let flow = await flows.textInputWithActions.start({});

  // Submit action should fail with invalid email
  flow = await flows.textInputWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    { email: "invalid" },
    "submit"
  );

  expect((flow.steps[0].ui as any)?.hasValidationErrors).toBe(true);
  expect((flow.steps[0].ui as any)?.content[0].validationError).toEqual(
    "Invalid email format"
  );
  expect(flow.steps[0].status).toBe("PENDING");

  // Submit action should pass with valid email
  flow = await flows.textInputWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    { email: "test@example.com" },
    "submit"
  );

  expect(flow.steps[0].status).toBe("COMPLETED");
  expect(flow.status).toBe("COMPLETED");
});

test("flows - text input with actions - draft allows any value", async () => {
  let flow = await flows.textInputWithActions.start({});

  // Draft action should pass even with invalid email
  flow = await flows.textInputWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    { email: "not-an-email" },
    "draft"
  );

  expect(flow.steps[0].status).toBe("COMPLETED");
  expect(flow.status).toBe("COMPLETED");
});

test("flows - number input with actions - buy requires minimum 1", async () => {
  let flow = await flows.numberInputWithActions.start({});

  // Buy action should fail with quantity 0
  flow = await flows.numberInputWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    { quantity: 0 },
    "buy"
  );

  expect((flow.steps[0].ui as any)?.hasValidationErrors).toBe(true);
  expect((flow.steps[0].ui as any)?.content[0].validationError).toEqual(
    "Must buy at least 1 item"
  );
  expect(flow.steps[0].status).toBe("PENDING");

  // Buy action should pass with quantity >= 1
  flow = await flows.numberInputWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    { quantity: 3 },
    "buy"
  );

  expect(flow.steps[0].status).toBe("COMPLETED");
  expect(flow.status).toBe("COMPLETED");
});

test("flows - number input with actions - reserve requires minimum 5", async () => {
  let flow = await flows.numberInputWithActions.start({});

  // Reserve action should fail with quantity < 5
  flow = await flows.numberInputWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    { quantity: 3 },
    "reserve"
  );

  expect((flow.steps[0].ui as any)?.hasValidationErrors).toBe(true);
  expect((flow.steps[0].ui as any)?.content[0].validationError).toEqual(
    "Must reserve at least 5 items"
  );
  expect(flow.steps[0].status).toBe("PENDING");

  // Reserve action should pass with quantity >= 5
  flow = await flows.numberInputWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    { quantity: 10 },
    "reserve"
  );

  expect(flow.steps[0].status).toBe("COMPLETED");
  expect(flow.status).toBe("COMPLETED");
});

test("flows - boolean input with actions - submit requires true", async () => {
  let flow = await flows.booleanInputWithActions.start({});

  // Submit action should fail when not agreed
  flow = await flows.booleanInputWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    { agreeToTerms: false },
    "submit"
  );

  expect((flow.steps[0].ui as any)?.hasValidationErrors).toBe(true);
  expect((flow.steps[0].ui as any)?.content[0].validationError).toEqual(
    "You must agree to the terms to submit"
  );
  expect(flow.steps[0].status).toBe("PENDING");

  // Submit action should pass when agreed
  flow = await flows.booleanInputWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    { agreeToTerms: true },
    "submit"
  );

  expect(flow.steps[0].status).toBe("COMPLETED");
  expect(flow.status).toBe("COMPLETED");
});

test("flows - boolean input with actions - draft allows any value", async () => {
  let flow = await flows.booleanInputWithActions.start({});

  // Draft action should pass even when not agreed
  flow = await flows.booleanInputWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    { agreeToTerms: false },
    "draft"
  );

  expect(flow.steps[0].status).toBe("COMPLETED");
  expect(flow.status).toBe("COMPLETED");
});

test("flows - date picker with actions - schedule requires future date", async () => {
  let flow = await flows.datePickerWithActions.start({});

  // Schedule action should fail with past date
  const yesterday = new Date();
  yesterday.setDate(yesterday.getDate() - 1);

  flow = await flows.datePickerWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    { startDate: yesterday },
    "schedule"
  );

  expect((flow.steps[0].ui as any)?.hasValidationErrors).toBe(true);
  expect((flow.steps[0].ui as any)?.content[0].validationError).toEqual(
    "Start date must be in the future for scheduling"
  );
  expect(flow.steps[0].status).toBe("PENDING");

  // Schedule action should pass with future date
  const tomorrow = new Date();
  tomorrow.setDate(tomorrow.getDate() + 1);

  flow = await flows.datePickerWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    { startDate: tomorrow },
    "schedule"
  );

  expect(flow.steps[0].status).toBe("COMPLETED");
  expect(flow.status).toBe("COMPLETED");
});

test("flows - date picker with actions - draft allows any date", async () => {
  let flow = await flows.datePickerWithActions.start({});

  // Draft action should pass even with past date
  const yesterday = new Date();
  yesterday.setDate(yesterday.getDate() - 1);

  flow = await flows.datePickerWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    { startDate: yesterday },
    "draft"
  );

  expect(flow.steps[0].status).toBe("COMPLETED");
  expect(flow.status).toBe("COMPLETED");
});

test("flows - iterator with actions - finalize requires 2+ items", async () => {
  let flow = await flows.iteratorWithActions.start({});

  // Finalize action should fail with less than 2 items
  flow = await flows.iteratorWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    { items: [{ itemName: "Item 1", price: 100 }] },
    "finalize"
  );

  expect((flow.steps[0].ui as any)?.hasValidationErrors).toBe(true);
  expect((flow.steps[0].ui as any)?.content[0].validationError).toEqual(
    "Must have at least 2 items to finalize"
  );
  expect(flow.steps[0].status).toBe("PENDING");

  // Finalize action should fail when prices are not positive
  flow = await flows.iteratorWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    {
      items: [
        { itemName: "Item 1", price: 0 },
        { itemName: "Item 2", price: 50 },
      ],
    },
    "finalize"
  );

  expect((flow.steps[0].ui as any)?.hasValidationErrors).toBe(true);
  expect(flow.steps[0].status).toBe("PENDING");

  // Finalize action should pass with valid data
  flow = await flows.iteratorWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    {
      items: [
        { itemName: "Item 1", price: 100 },
        { itemName: "Item 2", price: 50 },
      ],
    },
    "finalize"
  );

  expect(flow.steps[0].status).toBe("COMPLETED");
  expect(flow.status).toBe("COMPLETED");
});

test("flows - iterator with actions - save allows any data", async () => {
  let flow = await flows.iteratorWithActions.start({});

  // Save action should pass even with 1 item and negative price
  flow = await flows.iteratorWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    { items: [{ itemName: "Item 1", price: -10 }] },
    "save"
  );

  expect(flow.steps[0].status).toBe("COMPLETED");
  expect(flow.status).toBe("COMPLETED");
});

test("flows - dataGrid with actions - approve requires valid inventory", async () => {
  let flow = await flows.dataGridWithActions.start({});

  // Approve should fail with insufficient items
  flow = await flows.dataGridWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    { inventory: [{ id: "item-1", sku: "SKU001", quantity: 10, price: 50 }] },
    "approve"
  );

  expect((flow.steps[0].ui as any)?.hasValidationErrors).toBe(true);
  expect((flow.steps[0].ui as any)?.content[0].validationError).toEqual(
    "Must have at least 2 items to approve"
  );
  expect(flow.steps[0].status).toBe("PENDING");

  // Approve should fail with zero quantity
  flow = await flows.dataGridWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    {
      inventory: [
        { id: "item-1", sku: "SKU001", quantity: 0, price: 50 },
        { id: "item-2", sku: "SKU002", quantity: 5, price: 50 },
      ],
    },
    "approve"
  );

  expect((flow.steps[0].ui as any)?.hasValidationErrors).toBe(true);
  expect((flow.steps[0].ui as any)?.content[0].validationError).toEqual(
    "All items must have positive quantities when approving"
  );
  expect(flow.steps[0].status).toBe("PENDING");

  // Approve should fail with total value < $100
  flow = await flows.dataGridWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    {
      inventory: [
        { id: "item-1", sku: "SKU001", quantity: 1, price: 20 },
        { id: "item-2", sku: "SKU002", quantity: 1, price: 30 },
      ],
    },
    "approve"
  );

  expect((flow.steps[0].ui as any)?.hasValidationErrors).toBe(true);
  expect((flow.steps[0].ui as any)?.content[0].validationError).toEqual(
    "Total value must be at least $100 to approve"
  );
  expect(flow.steps[0].status).toBe("PENDING");

  // Approve should pass with valid data
  flow = await flows.dataGridWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    {
      inventory: [
        { id: "item-1", sku: "SKU001", quantity: 2, price: 30 },
        { id: "item-2", sku: "SKU002", quantity: 2, price: 25 },
      ],
    },
    "approve"
  );

  expect(flow.steps[0].status).toBe("COMPLETED");
  expect(flow.status).toBe("COMPLETED");
});

test("flows - dataGrid with actions - draft allows any data", async () => {
  let flow = await flows.dataGridWithActions.start({});

  // Draft should allow empty data
  flow = await flows.dataGridWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    { inventory: [] },
    "draft"
  );

  expect(flow.steps[0].status).toBe("COMPLETED");
  expect(flow.status).toBe("COMPLETED");
});

test("flows - pickList with actions - complete requires all items picked", async () => {
  let flow = await flows.pickListWithActions.start({});

  // Complete should fail when not fully picked
  flow = await flows.pickListWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    {
      items: {
        items: [
          {
            id: "prod-1",
            quantity: 5,
            targetQuantity: 10,
            scannedBarcodes: [],
          },
          { id: "prod-2", quantity: 5, targetQuantity: 5, scannedBarcodes: [] },
          { id: "prod-3", quantity: 3, targetQuantity: 3, scannedBarcodes: [] },
        ],
      },
    },
    "complete"
  );

  expect((flow.steps[0].ui as any)?.hasValidationErrors).toBe(true);
  expect((flow.steps[0].ui as any)?.content[0].validationError).toEqual(
    "All items must be fully picked to complete"
  );
  expect(flow.steps[0].status).toBe("PENDING");

  // Complete should pass when all items fully picked
  flow = await flows.pickListWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    {
      items: {
        items: [
          {
            id: "prod-1",
            quantity: 10,
            targetQuantity: 10,
            scannedBarcodes: [],
          },
          { id: "prod-2", quantity: 5, targetQuantity: 5, scannedBarcodes: [] },
          { id: "prod-3", quantity: 3, targetQuantity: 3, scannedBarcodes: [] },
        ],
      },
    },
    "complete"
  );

  expect(flow.steps[0].status).toBe("COMPLETED");
  expect(flow.status).toBe("COMPLETED");
});

test("flows - pickList with actions - partial requires at least one item", async () => {
  let flow = await flows.pickListWithActions.start({});

  // Partial should fail with no items picked
  flow = await flows.pickListWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    {
      items: {
        items: [
          {
            id: "prod-1",
            quantity: 0,
            targetQuantity: 10,
            scannedBarcodes: [],
          },
          { id: "prod-2", quantity: 0, targetQuantity: 5, scannedBarcodes: [] },
          { id: "prod-3", quantity: 0, targetQuantity: 3, scannedBarcodes: [] },
        ],
      },
    },
    "partial"
  );

  expect((flow.steps[0].ui as any)?.hasValidationErrors).toBe(true);
  expect((flow.steps[0].ui as any)?.content[0].validationError).toEqual(
    "At least one item must be picked for partial completion"
  );
  expect(flow.steps[0].status).toBe("PENDING");

  // Partial should pass with at least one item picked
  flow = await flows.pickListWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    {
      items: {
        items: [
          {
            id: "prod-1",
            quantity: 2,
            targetQuantity: 10,
            scannedBarcodes: [],
          },
          { id: "prod-2", quantity: 0, targetQuantity: 5, scannedBarcodes: [] },
          { id: "prod-3", quantity: 0, targetQuantity: 3, scannedBarcodes: [] },
        ],
      },
    },
    "partial"
  );

  expect(flow.steps[0].status).toBe("COMPLETED");
  expect(flow.status).toBe("COMPLETED");
});

test("flows - pickList with actions - cancel allows any state", async () => {
  let flow = await flows.pickListWithActions.start({});

  // Cancel should allow any state
  flow = await flows.pickListWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    {
      items: {
        items: [
          {
            id: "prod-1",
            quantity: 0,
            targetQuantity: 10,
            scannedBarcodes: [],
          },
          { id: "prod-2", quantity: 0, targetQuantity: 5, scannedBarcodes: [] },
          { id: "prod-3", quantity: 0, targetQuantity: 3, scannedBarcodes: [] },
        ],
      },
    },
    "cancel"
  );

  expect(flow.steps[0].status).toBe("COMPLETED");
  expect(flow.status).toBe("COMPLETED");
});

test("flows - scan with actions - single mode verify validates format", async () => {
  let flow = await flows.scanWithActions.start({});

  // Verify should fail without PROD- prefix
  flow = await flows.scanWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    { productCode: "ABC" },
    "verify"
  );

  expect((flow.steps[0].ui as any)?.hasValidationErrors).toBe(true);
  expect((flow.steps[0].ui as any)?.content[0].validationError).toEqual(
    "Product code must start with 'PROD-' for verification"
  );
  expect(flow.steps[0].status).toBe("PENDING");

  // Verify should fail when too short
  flow = await flows.scanWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    { productCode: "PROD-123" },
    "verify"
  );

  expect((flow.steps[0].ui as any)?.hasValidationErrors).toBe(true);
  expect((flow.steps[0].ui as any)?.content[0].validationError).toEqual(
    "Product code must be at least 10 characters for verification"
  );
  expect(flow.steps[0].status).toBe("PENDING");

  // Verify should pass with valid code
  flow = await flows.scanWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    { productCode: "PROD-12345" },
    "verify"
  );

  expect(flow.steps[0].status).toBe("COMPLETED");
});

test("flows - scan with actions - single mode lookup is lenient", async () => {
  let flow = await flows.scanWithActions.start({});

  // Lookup should pass with any non-empty value
  flow = await flows.scanWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    { productCode: "ABC123" },
    "lookup"
  );

  expect(flow.steps[0].status).toBe("COMPLETED");
});

test("flows - scan with actions - multi mode process validates", async () => {
  let flow = await flows.scanWithActions.start({});

  // Complete first page
  flow = await flows.scanWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    { productCode: "PROD-12345" },
    "verify"
  );

  // Process should fail with less than 3 items
  flow = await flows.scanWithActions.putStepValues(
    flow.id,
    flow.steps[1].id,
    { barcodes: ["123", "456"] },
    "process"
  );

  expect((flow.steps[1].ui as any)?.hasValidationErrors).toBe(true);
  expect((flow.steps[1].ui as any)?.content[0].validationError).toEqual(
    "Must scan at least 3 items to process"
  );
  expect(flow.steps[1].status).toBe("PENDING");

  // Process should fail with non-numeric codes
  flow = await flows.scanWithActions.putStepValues(
    flow.id,
    flow.steps[1].id,
    { barcodes: ["123", "456", "ABC"] },
    "process"
  );

  expect((flow.steps[1].ui as any)?.hasValidationErrors).toBe(true);
  expect((flow.steps[1].ui as any)?.content[0].validationError).toEqual(
    "All barcodes must be numeric for processing"
  );
  expect(flow.steps[1].status).toBe("PENDING");

  // Process should pass with valid data
  flow = await flows.scanWithActions.putStepValues(
    flow.id,
    flow.steps[1].id,
    { barcodes: ["123", "456", "789"] },
    "process"
  );

  expect(flow.steps[1].status).toBe("COMPLETED");
});

test("flows - scan with actions - quantity tracking checkout validates", async () => {
  let flow = await flows.scanWithActions.start({});

  // Complete first two pages
  flow = await flows.scanWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    { productCode: "PROD-12345" },
    "verify"
  );

  flow = await flows.scanWithActions.putStepValues(
    flow.id,
    flow.steps[1].id,
    { barcodes: ["123", "456", "789"] },
    "process"
  );

  // Checkout should fail with total quantity < 5
  flow = await flows.scanWithActions.putStepValues(
    flow.id,
    flow.steps[2].id,
    {
      items: [
        { value: "ITEM1", quantity: 2 },
        { value: "ITEM2", quantity: 1 },
      ],
    },
    "checkout"
  );

  expect((flow.steps[2].ui as any)?.hasValidationErrors).toBe(true);
  expect((flow.steps[2].ui as any)?.content[0].validationError).toEqual(
    "Total quantity must be at least 5 for checkout"
  );
  expect(flow.steps[2].status).toBe("PENDING");

  // Checkout should fail if any item has quantity < 2
  flow = await flows.scanWithActions.putStepValues(
    flow.id,
    flow.steps[2].id,
    {
      items: [
        { value: "ITEM1", quantity: 4 },
        { value: "ITEM2", quantity: 1 },
      ],
    },
    "checkout"
  );

  expect((flow.steps[2].ui as any)?.hasValidationErrors).toBe(true);
  expect((flow.steps[2].ui as any)?.content[0].validationError).toEqual(
    "Each item must have quantity of at least 2 for checkout"
  );
  expect(flow.steps[2].status).toBe("PENDING");

  // Checkout should pass with valid quantities
  flow = await flows.scanWithActions.putStepValues(
    flow.id,
    flow.steps[2].id,
    {
      items: [
        { value: "ITEM1", quantity: 3 },
        { value: "ITEM2", quantity: 2 },
      ],
    },
    "checkout"
  );

  expect(flow.steps[2].status).toBe("COMPLETED");
  expect(flow.status).toBe("COMPLETED");
});

test("flows - scan with actions - quantity tracking continue is lenient", async () => {
  let flow = await flows.scanWithActions.start({});

  // Complete first two pages
  flow = await flows.scanWithActions.putStepValues(
    flow.id,
    flow.steps[0].id,
    { productCode: "ABC" },
    "lookup"
  );

  flow = await flows.scanWithActions.putStepValues(
    flow.id,
    flow.steps[1].id,
    { barcodes: ["123"] },
    "save"
  );

  // Continue should pass with minimal data
  flow = await flows.scanWithActions.putStepValues(
    flow.id,
    flow.steps[2].id,
    { items: [{ value: "ITEM1", quantity: 1 }] },
    "continue"
  );

  expect(flow.steps[2].status).toBe("COMPLETED");
  expect(flow.status).toBe("COMPLETED");
});

test("flows - file inputs flow", async () => {
  let flow = await flows.fileInput.start({});
  expect(flow).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "AWAITING_INPUT",
    name: "FileInput",
    startedBy: null,
    input: {},
    error: null,
    data: null,
    steps: [
      {
        id: expect.any(String),
        name: "file input page",
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
              __type: "ui.input.file",
              label: "Avatar",
              disabled: false,
              helpText: "A nice photo of yourself",
              name: "avatar",
              optional: true,
            },
            {
              __type: "ui.input.file",
              disabled: false,
              label: "Passport",
              name: "passport",
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
      title: "File input",
    },
  });

  const avatarCallbackResponse = await flows.fileInput.callback(
    flow.id,
    flow.steps[0].id,
    "avatar",
    "getPresignedUploadURL",
    {}
  );
  expect(avatarCallbackResponse).toEqual({
    key: expect.any(String),
    url: expect.any(String),
  });

  const passportCallbackResponse = await flows.fileInput.callback(
    flow.id,
    flow.steps[0].id,
    "passport",
    "getPresignedUploadURL",
    { key: "existing-file-key" }
  );
  expect(passportCallbackResponse).toEqual({
    key: "existing-file-key",
    url: expect.any(String),
  });

  // client would now upload to the `passportCallbackResponse.url` and `avatarCallbackResponse.url` the
  // files, and then submit the page

  flow = await flows.fileInput.putStepValues(flow.id, flow.steps[0].id, {
    avatar: {
      key: avatarCallbackResponse.key,
      filename: "my-avatar.png",
      contentType: "image/png",
    },
    passport: {
      key: "existing-file-key",
      filename: "my-passport.pdf",
      contentType: "application/pdf",
    },
  });

  expect(flow.status).toBe("COMPLETED");
  expect(flow.steps[0].status).toBe("COMPLETED");
  expect(flow.steps[0].value).toEqual({
    avatar: {
      key: expect.any(String),
      filename: "my-avatar.png",
      contentType: "image/png",
    },
    passport: {
      key: "existing-file-key",
      filename: "my-passport.pdf",
      contentType: "application/pdf",
    },
  });
});
