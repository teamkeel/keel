import { actions, resetDatabase, models } from "@teamkeel/testing";
import { beforeEach, expect, test } from "vitest";

beforeEach(resetDatabase);

test("set with env var", async () => {
  const person = await actions.createPerson({});

  expect(person.name).toEqual("Pedro");
});

test("permissions with env var", async () => {
  const person = await actions.createPerson({});

  await expect(
    actions.getPedro({ id: person.id })
  ).not.toHaveAuthorizationError();

  await expect(actions.getBob({ id: person.id })).toHaveAuthorizationError();
});

test("where with env var", async () => {
  await models.person.create({ name: "Pedro" });
  await models.person.create({ name: "Bob" });

  const pedros = await actions.listPedros();
  expect(pedros.results.length).toEqual(1);
  expect(pedros.results[0].name).toEqual("Pedro");
});
