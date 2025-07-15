import { resetDatabase, models } from "@teamkeel/testing";
import { beforeEach, expect, test } from "vitest";

beforeEach(resetDatabase);

test("flows - iterator element", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  let { status, body } = await startFlow({
    name: "Iterator",
    token,
    body: {},
  });

  expect(status).toEqual(200);
  expect(body).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "AWAITING_INPUT",
    name: "Iterator",
    startedBy: expect.any(String),
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
        startTime: expect.any(String),
        endTime: null,
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
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
              max: 5,
              min: 1,
              name: "my iterator",
            },
          ],
          hasValidationErrors: false,
        },
      },
    ],
    createdAt: expect.any(String),
    updatedAt: expect.any(String),
    config: {
      title: "Iterator",
    },
  });

   // Provide the values for the pending UI step
   ({ status, body } = await putStepValues({
    name: "Iterator",
    runId: body.id,
    stepId: body.steps[0].id,
    token,
    values: {
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
    action: null,
  }));
  expect(status).toEqual(200);
  expect(body).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "COMPLETED",
    name: "Iterator",
    startedBy: expect.any(String),
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
      title: "Iterator",
    },
  });
});


test("flows - iterator element - too few items in iterator", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  let { status, body } = await startFlow({
    name: "Iterator",
    token,
    body: {},
  });

  expect(status).toEqual(200);

   // Provide the values for the pending UI step
   ({ status, body } = await putStepValues({
    name: "Iterator",
    runId: body.id,
    stepId: body.steps[0].id,
    token,
    values: {
      "my iterator": [

      ],
    },
    action: null,
  }));
  expect(status).toEqual(200);
  expect(body).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "COMPLETED",
    name: "Iterator",
    startedBy: expect.any(String),
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
      title: "Iterator",
    },
  });
});


test.only("flows - iterator element - iterator has element with failed validation", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  let { status, body } = await startFlow({
    name: "Iterator",
    token,
    body: {},
  });

  expect(status).toEqual(200);

   // Provide the values for the pending UI step
   ({ status, body } = await putStepValues({
    name: "Iterator",
    runId: body.id,
    stepId: body.steps[0].id,
    token,
    values: {
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
           sku: "PANTS",
           quantity: 3,
         },
       ],
       
    },
    action: null,
  }));
  expect(status).toEqual(200);
  expect(body).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "AWAITING_INPUT",
    name: "Iterator",
    startedBy: expect.any(String),
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
        startTime: expect.any(String),
        endTime: null,
        createdAt: expect.any(String),
        updatedAt: expect.any(String),
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
              max: 5,
              min: 1,
              name: "my iterator",
            },
          ],
          hasValidationErrors: false,
        },
      },
    ],
    createdAt: expect.any(String),
    updatedAt: expect.any(String),
    config: {
      title: "Iterator",
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
