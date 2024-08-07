import { actions, resetDatabase, models } from "@teamkeel/testing";
import { beforeEach, expect, test } from "vitest";

beforeEach(resetDatabase);

test("set with http header", async () => {
  const response = await fetch(
    process.env.KEEL_TESTING_ACTIONS_API_URL + "/createPerson",
    {
      body: "{}",
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Person-Name": "Pedro",
      },
    }
  );

  expect(response.status).toEqual(200);
  const data = await response.json();
  expect(data?.name).toEqual("Pedro");
});

test("set with http header camel case in schema", async () => {
  const response = await fetch(
    process.env.KEEL_TESTING_ACTIONS_API_URL + "/createPerson",
    {
      body: "{}",
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Person-Name": "Pedro",
      },
    }
  );

  expect(response.status).toEqual(200);
  const data = await response.json();
  expect(data?.name).toEqual("Pedro");
});

test("set with http header - different casing but matching", async () => {
  const response = await fetch(
    process.env.KEEL_TESTING_ACTIONS_API_URL + "/createPerson",
    {
      body: "{}",
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "person-name": "Pedro",
      },
    }
  );

  expect(response.status).toEqual(200);
  const data = await response.json();
  expect(data?.name).toEqual("Pedro");
});

test("set with http header - not matching", async () => {
  const response = await fetch(
    process.env.KEEL_TESTING_ACTIONS_API_URL + "/createPerson",
    {
      body: "{}",
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        personname: "Pedro",
      },
    }
  );

  expect(response.status).toEqual(200);
  const data = await response.json();
  expect(data?.name).toEqual("");
});

test("set with http header - not matching", async () => {
  const response = await fetch(
    process.env.KEEL_TESTING_ACTIONS_API_URL + "/createPerson",
    {
      body: "{}",
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        person_name: "Pedro",
      },
    }
  );

  expect(response.status).toEqual(200);
  const data = await response.json();
  expect(data?.name).toEqual("");
});

test("create with http header in hook function", async () => {
  const response = await fetch(
    process.env.KEEL_TESTING_ACTIONS_API_URL + "/createPersonUsingHook",
    {
      body: "{}",
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Person-Name": "Pedro",
      },
    }
  );

  expect(response.status).toEqual(200);
  const data = await response.json();
  expect(data?.name).toEqual("Pedro");
});

test("create with http header in custom function", async () => {
  const response = await fetch(
    process.env.KEEL_TESTING_ACTIONS_API_URL + "/writePersonUsingCustomFunc",
    {
      body: "{}",
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Person-Name": "Pedro",
      },
    }
  );

  expect(response.status).toEqual(200);
  const data = await response.json();
  expect(data?.name).toEqual("Pedro");
});

test("permissions with http header", async () => {
  const response = await fetch(
    process.env.KEEL_TESTING_ACTIONS_API_URL + "/createPerson",
    {
      body: "{}",
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Person-Name": "Pedro",
      },
    }
  );

  expect(response.status).toEqual(200);
  const person = await response.json();

  const getPedro = await fetch(
    process.env.KEEL_TESTING_ACTIONS_API_URL + "/getPedro",
    {
      body: `{ "id": "${person.id}"}`,
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Person-Name": "Pedro",
      },
    }
  );

  expect(getPedro.status).toEqual(200);

  const getBob = await fetch(
    process.env.KEEL_TESTING_ACTIONS_API_URL + "/getBob",
    {
      body: `{ "id": "${person.id}"}`,
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Person-Name": "Pedro",
      },
    }
  );

  expect(getBob.status).toEqual(403);
});

test("where with http header", async () => {
  await models.person.create({ name: "Pedro" });
  await models.person.create({ name: "Bob" });

  const listPedros = await fetch(
    process.env.KEEL_TESTING_ACTIONS_API_URL + "/listPedros",
    {
      body: "{}",
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Person-Name": "Pedro",
      },
    }
  );

  expect(listPedros.status).toEqual(200);

  const pedros = await listPedros.json();
  expect(pedros.results.length).toEqual(1);
  expect(pedros.results[0].name).toEqual("Pedro");
});
