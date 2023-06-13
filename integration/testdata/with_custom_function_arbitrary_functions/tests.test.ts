import { actions, models, resetDatabase } from "@teamkeel/testing";
import { Person } from "@teamkeel/sdk";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

test("arbitrary read function with inline inputs", async () => {
  await models.person.create({
    name: "Keelson",
  });
  await models.person.create({
    name: "Weaveton",
  });
  await models.person.create({
    name: "Keeler",
  });

  const result = await actions.countName({ name: "Keelson" });
  expect(result.count).toEqual(1);
});

test("arbitrary read function with message input", async () => {
  await models.person.create({
    name: "Keelson",
  });
  await models.person.create({
    name: "Weaveton",
  });
  await models.person.create({
    name: "Keeler",
  });

  const result = await actions.countNameAdvanced({
    startsWith: "Kee",
    contains: "e",
    endsWith: "r",
  });
  expect(result.count).toEqual(1);
});

test("arbitrary write function with inline inputs", async () => {
  const result1 = await actions.createAndCount({ name: "Keelson" });
  expect(result1.count).toEqual(1);

  const result2 = await actions.createAndCount({ name: "Keelson" });
  expect(result2.count).toEqual(2);
});

test("arbitrary write function with message input", async () => {
  const result1 = await actions.createManyAndCount({
    names: ["Keelson", "Weaveton"],
  });
  expect(result1.count).toEqual(2);

  const result2 = await actions.createManyAndCount({
    names: ["Keelson", "Weaveton"],
  });
  expect(result2.count).toEqual(4);
});

const anyTypeFixtures = [
  {
    description: "Boolean",
    value: true,
  },
  {
    description: "String",
    value: "hello world",
  },
  {
    description: "Number",
    value: 123,
  },
  {
    description: "Number Array",
    value: [123, 234],
  },
  {
    description: "String Array",
    value: ["one", "two", "three"],
  },
  {
    description: "Object",
    value: {
      name: "123",
    },
  },
  {
    description: "Array of Objects",
    value: [
      {
        name: "123",
      },
      {
        name: "234",
      },
    ],
  },
];

anyTypeFixtures.forEach(({ description, value }) => {
  test(`Message types with 'Any' type - ${description}`, async () => {
    const result = await actions.customSearch(value);

    expect(result).toEqual(value);
  });
});

test("Message types with fields of 'Any' type", async () => {
  await models.person.create({
    name: "Keelson",
  });
  await models.person.create({
    name: "Weaveton",
  });
  await models.person.create({
    name: "Keeler",
  });

  // params is any type
  const params = {
    names: ["Keelson", "Weaveton", "Keeler"],
  };

  const result = await actions.customPersonSearch({ params });

  // result.people is also any in the return type
  expect(result.people.map((p) => p.name).sort()).toEqual(params.names.sort());
});

test("No inputs", async () => {
  const result = await actions.noInputs();
  expect(result.success).toEqual(true);
});

test("Message with field of type Model", async () => {
  const person_1: Person = {
    id: "234",
    createdAt: new Date(),
    updatedAt: new Date(),
    name: "Adam",
  };

  const person_2: Person = {
    id: "123",
    createdAt: new Date(),
    updatedAt: new Date(),
    name: "Bob",
  };

  const peopleToUpload = [person_1, person_2];

  const result = await actions.bulkPersonUpload({ people: peopleToUpload });

  expect(result.people.map((p) => p.id)).toEqual(
    peopleToUpload.map((p) => p.id)
  );
});
