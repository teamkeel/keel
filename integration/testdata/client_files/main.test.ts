import { actions, models, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach, beforeAll } from "vitest";
import { APIClient } from "./keelClient";

var client: APIClient;

beforeEach(() => {
  client = new APIClient({ baseUrl: process.env.KEEL_TESTING_CLIENT_API_URL! });
});

beforeEach(resetDatabase);

test("client - create with file", async () => {
  const fileContents = "hello";
  const dataUrl = `data:text/plain;name=my-file.txt;base64,${Buffer.from(
    fileContents
  ).toString("base64")}`;

  const result = await client.api.mutations.createAccount({
    name: "Keelson",
    data: dataUrl,
  });

  expect(result.data?.data.contentType).toEqual("text/plain");
  expect(result.data?.data.filename).toEqual("my-file.txt");
  expect(result.data?.data.size).toEqual(5);

  const response = await fetch(new URL(result.data!.data.url));
  const buffer = Buffer.from(await response.arrayBuffer());
  const contents = buffer.toString("utf-8");
  expect(contents).toEqual(fileContents);
});

test("client - list with file", async () => {
  const fileContents = "hello";
  const dataUrl = `data:text/plain;name=my-file.txt;base64,${Buffer.from(
    fileContents
  ).toString("base64")}`;

  await client.api.mutations.createAccount({
    name: "Keelson",
    data: dataUrl,
  });

  const result = await client.api.queries.listAccounts();

  expect(result.data?.results[0].data.contentType).toEqual("text/plain");
  expect(result.data?.results[0].data.filename).toEqual("my-file.txt");
  expect(result.data?.results[0].data.size).toEqual(5);

  const response = await fetch(new URL(result.data!.results[0].data.url));
  const buffer = Buffer.from(await response.arrayBuffer());
  const contents = buffer.toString("utf-8");
  expect(contents).toEqual(fileContents);
});

test("client - write action with file", async () => {
  const fileContents = "hello";
  const dataUrl = `data:text/plain;name=my-file.txt;base64,${Buffer.from(
    fileContents
  ).toString("base64")}`;

  const result = await client.api.mutations.writeAccounts({
    csv: dataUrl,
  });

  expect(result.data?.csv.contentType).toEqual("text/plain");
  expect(result.data?.csv.filename).toEqual("my-file.txt");
  expect(result.data?.csv.size).toEqual(5);

  const response = await fetch(new URL(result.data!.csv.url));
  const buffer = Buffer.from(await response.arrayBuffer());
  const contents = buffer.toString("utf-8");
  expect(contents).toEqual(fileContents);
});
