import { actions, models, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

test("create with id", async () => {
    const person = await actions.createPerson({ id: "1", name: "Keelson" });

    expect(person.id).toEqual("1");

    const getPerson = await actions.getPerson({ id: "1"});

    expect(person.id).toEqual("1");
});


test("create with @set id", async () => {
  const person = await actions.createPerson({ id: "1", name: "Keelson" });

  expect(person.id).toEqual("1");

  const getPerson = await actions.getPerson({ id: "1"});

  expect(person.id).toEqual("1");
});


test("create with empty id", async () => {
  const person = await actions.createPerson({ id: "", name: "Keelson" });

  expect(person.id).toEqual("");

 // const person2 = await actions.createPerson({ id: "", name: "Keelson" });

  // const getPerson = await actions.getPerson({ id: "1"});

  // expect(person.id).toEqual("1");
});

