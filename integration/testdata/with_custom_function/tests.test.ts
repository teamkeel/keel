import { actions, models, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

test("creating a person", async () => {
  const person = await actions.createPerson({
    name: "foo",
    gender: "female",
    niNumber: "282",
  });

  expect(person.name).toEqual("foo");
});

test("creating a person with identity", async () => {
  const { token: token } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "user@keel.xyz",
      password: "1234",
    },
  });

  const identity = await models.identity.findOne({ email: "user@keel.xyz" });
  expect(identity).not.toBeNull();

  const person = await actions
    .withAuthToken(token)
    .createPersonWithContextInfo({
      name: "foo",
      gender: "female",
      niNumber: "771",
    });

  expect(person.name).toEqual("user@keel.xyz");
  expect(person.identityId).toEqual(identity?.id);
  expect(person.ctxNow).not.toBeNull();
});

test("creating a person without identity", async () => {
  const person = await actions.createPersonWithContextInfo({
    name: "foo",
    gender: "female",
    niNumber: "771",
  });

  expect(person.name).toEqual("none");
  expect(person.identityId).toBeNull();
  expect(person.ctxNow).not.toBeNull();
});

test("fetching a person by id", async () => {
  const person = await models.person.create({
    name: "bar",
    gender: "male",
    niNumber: "123",
  });
  const fetchedPerson = await actions.getPerson({ id: person.id });

  expect(person.id).toEqual(fetchedPerson!.id);
});

test("fetching person by additional unique field (not PK)", async () => {
  const person = await models.person.create({
    name: "bar",
    gender: "male",
    niNumber: "333",
  });

  const fetchedPerson = await actions.getPersonByNiNumber({ niNumber: "333" });

  expect(person.id).toEqual(fetchedPerson!.id);
});

test("listing with constraints", async () => {
  await models.person.create({ name: "fred", gender: "male", niNumber: "000" });
  const x11 = await models.person.create({
    name: "X11",
    gender: "alien",
    niNumber: "920",
  });
  const x22 = await models.person.create({
    name: "X22",
    gender: "alien",
    niNumber: "902",
  });

  const resp = await actions.listPeople({
    where: {
      gender: { equals: "alien" },
    },
  });

  const alienNames = resp.results.map((a) => a.name);

  expect(alienNames.sort()).toEqual([x11.name, x22.name].sort());
});

test("updating", async () => {
  const person = await models.person.create({
    name: "fred",
    gender: "male",
    niNumber: "678",
  });

  const updatedPerson = await actions.updatePerson({
    where: { id: person.id },
    values: { name: "paul", gender: "non-binary", niNumber: "789" },
  });

  expect(updatedPerson.name).toEqual("paul");
  expect(updatedPerson.gender).toEqual("non-binary");
  expect(updatedPerson.niNumber).toEqual("789");
  expect(updatedPerson.id).toEqual(person.id);
});

test("updating non existent record", async () => {
  await expect(
    actions.updatePerson({
      where: {
        id: "fake-id",
      },
      values: {
        name: "something",
        gender: "non-binary",
        niNumber: "1929",
      },
    })
  ).toHaveError({
    code: "ERR_RECORD_NOT_FOUND",
    message: "record not found",
  });
});

test("deleting", async () => {
  const person = await models.person.create({
    name: "fred",
    gender: "male",
    niNumber: "678",
  });

  const deletedId = await actions.deletePerson({ id: person.id });

  expect(deletedId).toEqual(person.id);
});

test("deleting non existent record", async () => {
  await expect(
    actions.deletePerson({
      id: "fake-id",
    })
  ).toHaveError({
    code: "ERR_RECORD_NOT_FOUND",
    message: "record not found",
  });
});

test("uniqueness constraint violation", async () => {
  const person = await models.person.create({
    name: "adam",
    niNumber: "123",
    gender: "non-binary",
  });

  await expect(
    actions.createPerson({
      name: "bob",
      niNumber: person.niNumber,
      gender: "non-binary",
    })
  ).toHaveError({
    code: "ERR_INVALID_INPUT",
    message: "the value for the unique field 'niNumber' must be unique",
  });
});

test("null value in foreign key column", async () => {
  await expect(actions.createProfileWithNullPerson({})).toHaveError({
    code: "ERR_INVALID_INPUT",
    message: "field 'personId' cannot be null",
  });
});

test("unrecognised value in foreign key column", async () => {
  await expect(
    actions.createProfile({
      person: { id: "missing-id" },
    })
  ).toHaveError({
    code: "ERR_INVALID_INPUT",
    message: "the record referenced in field 'personId' does not exist",
  });
});

test("relationships - findOne using multiple hasMany relationships", async () => {
  const publisher1 = await models.publisher.create({
    name: "Macmillan",
  });
  const publisher2 = await models.publisher.create({
    name: "Harper Collins",
  });
  const author1 = await models.author.create({
    publisherId: publisher1.id,
    name: "Philip K. Dick",
  });
  const author2 = await models.author.create({
    publisherId: publisher2.id,
    name: "Charles Dickens",
  });
  await models.book.create({
    title: "The Man In the High Castle",
    authorId: author1.id,
  });
  const oliverTwist = await models.book.create({
    title: "Oliver Twist",
    authorId: author2.id,
  });
  const res = await actions.getPublisherByBook({
    bookId: oliverTwist.id,
  });

  expect(res).not.toBe(null);
  expect(res!.id).toBe(publisher2.id);

  const res2 = await actions.getPublisher({
    id: publisher2.id,
  });

  expect(res2).not.toBe(null);
  expect(res2!.id).toBe(publisher2.id);
});

test("relationships - findMany using multiple belongsTo relationships", async () => {
  const publisher1 = await models.publisher.create({
    name: "Macmillan",
  });
  const publisher2 = await models.publisher.create({
    name: "Harper Collins",
  });
  const author1 = await models.author.create({
    publisherId: publisher1.id,
    name: "Philip K. Dick",
  });
  const author2 = await models.author.create({
    publisherId: publisher2.id,
    name: "Charles Dickens",
  });
  await models.book.create({
    title: "The Man In the High Castle",
    authorId: author1.id,
  });
  await models.book.create({
    title: "A Christmas Carol",
    authorId: author2.id,
  });
  await models.book.create({
    title: "Oliver Twist",
    authorId: author2.id,
  });
  const res = await actions.listBooksByPublisherName({
    where: {
      author: { publisher: { name: { equals: "Harper Collins" } } },
    },
  });
  expect(res.results.length).toBe(2);
  expect(res.results.map((x) => x.title).sort()).toEqual([
    "A Christmas Carol",
    "Oliver Twist",
  ]);
});

test("using an environment variable", async () => {
  const res = await actions.createPersonWithEnvVar({
    name: "adam",
    gender: "non-binary",
    niNumber: "123",
  });

  // see keelconfig.yaml
  expect(res.name).toEqual("test");
});

test("using a secret", async () => {
  const res = await actions.createPersonWithSecret({
    name: "dave",
    gender: "loD",
    niNumber: "JS70",
  });

  // secret value is set in integration/integration_test.go
  expect(res.name).toEqual("worf");
});

test("custom permissions - permitted action", async () => {
  const res = await actions.customPermission({
    name: "Adam",
    gender: "non-binary",
    niNumber: "123",
  });

  expect(res.name).toEqual("Adam");
});

test("custom permissions - unpermitted action", async () => {
  await expect(
    actions.customPermission({
      name: "Pete",
      gender: "non-binary",
      niNumber: "123",
    })
  ).rejects.toThrow("not authorized to access this action");

  // check there are no records in the db as the transaction should
  // have rolled back
  const records = await models.person.where({ name: "Pete" }).findMany();

  expect(records.length).toEqual(0);
});
