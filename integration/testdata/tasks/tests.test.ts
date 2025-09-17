import { resetDatabase, models } from "@teamkeel/testing";
import { beforeEach, expect, test } from "vitest";

beforeEach(resetDatabase);

test("tasks - create", async () => {
  const token = await getToken({ email: "admin@keel.xyz" });

  const resCreate = await createTask({ topic: "DispatchOrder", token: token });
  expect(resCreate.status).toBe(200);
  expect(resCreate.body).toEqual({
    createdAt: expect.any(String),
    id: expect.any(String),
    name: "DispatchOrder",
    status: "NEW",
    updatedAt: expect.any(String),
  });

  const res = await listTasks({ topic: "DispatchOrder", token: token });
  expect(res.status).toBe(200);
  expect(res.body.length).toBe(1);
  expect(res.body).toEqual([
    {
      createdAt: expect.any(String),
      id: expect.any(String),
      name: "DispatchOrder",
      status: "NEW",
      updatedAt: expect.any(String),
    },
  ]);
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

  const resCreate = await createTask({ topic: "DispatchOrder", token: token });
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

  const resCreate1 = await createTask({ topic: "DispatchOrder", token: token });
  expect(resCreate1.status).toBe(200);

  const resCreate2 = await createTask({ topic: "DispatchOrder", token: token });
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

  const resList = await listTasks({ topic: "DispatchOrder", token: token });
  expect(resList.status).toBe(200);
  expect(resList.body).toEqual([resCreate2.body, resNext.body]);
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

async function createTask({ topic, token }) {
  const url = `${process.env.KEEL_TESTING_API_URL}/topics/json/${topic}/tasks`;
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

async function listTasks({ topic, token }) {
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
