import { actions, models, resetDatabase } from "@teamkeel/testing";
import { Person, InlineFile } from "@teamkeel/sdk";
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
    height: null,
    photo: null,
  };

  const person_2: Person = {
    id: "123",
    createdAt: new Date(),
    updatedAt: new Date(),
    name: "Bob",
    height: null,
    photo: null,
  };

  const peopleToUpload = [person_1, person_2];

  const result = await actions.bulkPersonUpload({ people: peopleToUpload });

  expect(result.people.map((p) => p.id)).toEqual(
    peopleToUpload.map((p) => p.id)
  );
});

test("Create with decimal", async () => {
  const result = await actions.createPerson({
    name: "Keelio",
    height: 175.3,
  });

  expect(result.height).toEqual(175.3);
  expect(result.name).toEqual("Keelio");
});

test("File input handling", async () => {
  const dataUrl = `data:image/png;name=my-file.png;base64,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`;

  const result = await actions.fileInputHandling({
    file: dataUrl,
  });

  expect(result.filename).toEqual("my-file.png");
  expect(result.size).toEqual(2024);
  expect(result.contentType).toEqual("image/png");
  expect(result.key).toBeTruthy(); // A KSUID
});

test("createFromFile", async () => {
  const dataUrl = `data:image/png;name=my-file.png;base64,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`;

  const result = await actions.createFromFile({
    file: dataUrl,
  });

  expect(result.person.id).toBeTruthy();
  expect(result.person.name).toEqual("Pedro");
  expect(result.person.photo?.filename).toEqual("my-file.png");
  expect(result.person.photo?.size).toEqual(2024);
  expect(result.person.photo?.contentType).toEqual("image/png");
});

test("updateWithFile", async () => {
  const dataUrl = `data:image/png;name=my-file.png;base64,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`;

  const result = await actions.updateWithFile({
    file: dataUrl,
  });

  expect(result.person.id).toBeTruthy();
  expect(result.person.name).toEqual("Pedro");
  expect(result.person.photo?.filename).toEqual("my-file.png");
  expect(result.person.photo?.size).toEqual(2024);
  expect(result.person.photo?.contentType).toEqual("image/png");
});

test("updatePhoto", async () => {
  const keelio = await actions.createPerson({
    name: "Keelio",
    height: 175.3,
  });

  const dataUrl = `data:image/png;name=my-file.png;base64,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`;

  const result = await actions.updatePhoto({
    id: keelio.id,
    photo: dataUrl,
  });

  expect(result.person.id).toBeTruthy();
  expect(result.person.name).toEqual("Keelio");
  expect(result.person.photo?.filename).toEqual("my-file.png");
  expect(result.person.photo?.size).toEqual(2024);
  expect(result.person.photo?.contentType).toEqual("image/png");
});
