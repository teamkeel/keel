import { resetDatabase, models } from "@teamkeel/testing";
import { useDatabase } from "@teamkeel/sdk";
import { beforeEach, expect, test } from "vitest";

beforeEach(resetDatabase);

test("tasks - create", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({
    topic: "DispatchOrder",
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
    name: "DispatchOrder",
    status: "NEW",
    updatedAt: expect.any(String),
  });

  const res = await getTaskQueue({ topic: "DispatchOrder", token: token });
  expect(res.status).toBe(200);
  expect(res.body).toEqual([
    {
      createdAt: expect.any(String),
      id: expect.any(String),
      name: "DispatchOrder",
      status: "NEW",
      updatedAt: expect.any(String),
    },
  ]);

  const tasks = await useDatabase()
    .selectFrom("dispatch_order")
    .selectAll()
    .execute();

  expect(tasks).toEqual([
    {
      id: expect.any(String),
      orderDate: new Date(2025, 6, 15),
      shipByDate: new Date(2025, 6, 30), 
      createdAt: expect.any(Date),
      updatedAt: expect.any(Date),
      keelTaskId: resCreate.body.id,
    },
  ]);
});

test("tasks - create - no fields", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({
    topic: "NoFields",
    body: {},
    token: token,
  });
  expect(resCreate.status).toBe(200);
  expect(resCreate.body).toEqual({
    createdAt: expect.any(String),
    id: expect.any(String),
    name: "NoFields",
    status: "NEW",
    updatedAt: expect.any(String),
  });

  const res = await getTaskQueue({ topic: "NoFields", token: token });
  expect(res.status).toBe(200);
  expect(res.body).toEqual([
    {
      createdAt: expect.any(String),
      id: expect.any(String),
      name: "NoFields",
      status: "NEW",
      updatedAt: expect.any(String),
    },
  ]);
});

test("tasks - list", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const t1 = await createTask({
    topic: "DispatchOrder",
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
    topic: "DispatchOrder",
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
    topic: "DispatchOrder",
    body: {
      data: {
        orderDate: new Date(2025, 6, 14),
        shipByDate: new Date(2025, 6, 30),
      },
    },
    token: token,
  });
  expect(t3.status).toBe(200);
  const t4 = await createTask({
    topic: "DispatchOrder",
    body: {
      data: {
        orderDate: new Date(2025, 6, 10),
        shipByDate: new Date(2025, 6, 20),
      },
    },
    token: token,
  });
  expect(t4.status).toBe(200);
  const t5 = await createTask({
    topic: "DispatchOrder",
    body: {
      data: {
        orderDate: new Date(2025, 6, 10),
        shipByDate: new Date(2025, 6, 20),
      },
    },
    token: token,
  });
  expect(t5.status).toBe(200);

  const res = await getTaskQueue({ topic: "DispatchOrder", token: token });
  expect(res.status).toBe(200);
  expect(res.body).toEqual([t1.body, t5.body, t4.body, t3.body, t2.body]);
});

test("tasks - next - no tasks exist", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });
  const res = await nextTask({ topic: "DispatchOrder", token: token });
  expect(res.status).toBe(404);
  expect(res.body).toEqual({
    code: "ERR_RECORD_NOT_FOUND",
    message: "Not found",
  });
});

test("tasks - next - successfully assigned", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({
    topic: "DispatchOrder",
    body: {
      data: {
        orderDate: new Date(2025, 6, 9),
        shipByDate: new Date(2025, 6, 20),
      },
    },
    token: token,
  });
  expect(resCreate.status).toBe(200);

  const res = await nextTask({ topic: "DispatchOrder", token: token });
  expect(res.status).toBe(200);
  expect(res.body).toEqual({
    createdAt: expect.any(String),
    id: resCreate.body.id,
    name: "DispatchOrder",
    status: "ASSIGNED",
    updatedAt: expect.any(String),
    assignedTo: expect.any(String),
    assignedAt: expect.any(String),
  });
});

test("tasks - next - already assigned", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate1 = await createTask({
    topic: "DispatchOrder",
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
    topic: "DispatchOrder",
    body: {
      data: {
        orderDate: new Date(2025, 6, 9),
        shipByDate: new Date(2025, 6, 20),
      },
    },
    token: token,
  });
  expect(resCreate2.status).toBe(200);

  const resNext = await nextTask({ topic: "DispatchOrder", token: token });
  expect(resNext.status).toBe(200);
  expect(resNext.body).toEqual({
    createdAt: expect.any(String),
    id: resCreate1.body.id,
    name: "DispatchOrder",
    status: "ASSIGNED",
    updatedAt: expect.any(String),
    assignedTo: expect.any(String),
    assignedAt: expect.any(String),
  });

  const resNextAgain = await nextTask({ topic: "DispatchOrder", token: token });
  expect(resNextAgain.status).toBe(200);
  expect(resNextAgain.body).toEqual(resNext.body);

  const resList = await getTaskQueue({ topic: "DispatchOrder", token: token });
  expect(resList.status).toBe(200);
  expect(resList.body).toEqual([resNext.body, resCreate2.body]);
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
