import { resetDatabase, actions, models } from "@teamkeel/testing";
import { Sex } from "@teamkeel/sdk";
import { test, expect, beforeEach, describe } from "vitest";

beforeEach(resetDatabase);

describe("write hooks", () => {
  test("create.beforeWrite hook - mutating data", async () => {
    const identity = await models.identity.create({
      email: "adam@keel.xyz",
    });
    const name = "Adam";

    const record = await actions
      .withIdentity(identity)
      .createPersonWithBeforeWrite({
        title: name,
        sex: Sex.Male,
      });

    expect(record.title).toEqual(`Mr. ${name}`);
  });

  test("create.afterWrite hook - custom permissions check", async () => {
    const identity = await models.identity.create({
      email: "adam@keel.xyz",
    });

    await expect(
      actions.withIdentity(identity).createPersonWithAfterWrite({
        title: "Bob",
        sex: Sex.Male,
      })
    ).toHaveAuthorizationError();

    await expect(
      actions.withIdentity(identity).createPersonWithBeforeWrite({
        title: "Alice",
        sex: Sex.Female,
      })
    ).not.toHaveAuthorizationError();
  });

  test("update.beforeWrite hook - mutating data", async () => {
    const identity = await models.identity.create({
      email: "adam@keel.xyz",
    });
    const name = "Adam";

    const person = await models.person.create({
      sex: Sex.Male,
      title: name,
    });

    const record = await actions
      .withIdentity(identity)
      .updatePersonWithBeforeWrite({
        where: { id: person.id },
        values: {
          title: "Alice",
          sex: Sex.Female,
        },
      });

    expect(record.title).toEqual(`Ms. Alice`);
  });

  test("update.afterWrite hook - custom permissions check", async () => {
    const identity = await models.identity.create({
      email: "adam@keel.xyz",
    });
    const name = "Alice";

    const person = await models.person.create({
      sex: Sex.Female,
      title: name,
    });

    await expect(
      actions.withIdentity(identity).updatePersonWithAfterWrite({
        where: { id: person.id },
        values: {
          title: "Adam",
          sex: Sex.Male,
        },
      })
    ).toHaveAuthorizationError();
  });
});

describe("query hooks", () => {
  test("delete.beforeQuery", async () => {
    const identity = await models.identity.create({
      email: "adam@keel.xyz",
    });
    const person = await models.person.create({
      title: "adam",
      sex: Sex.Male,
    });

    await expect(
      actions.withIdentity(identity).deletePersonBeforeQuery({ id: person.id })
    ).toHaveError({
      message: "record not found",
    });

    const dbPerson = await models.person.findOne({
      id: person.id,
    });

    expect(dbPerson?.id).toEqual(person.id);
  });

  test("delete.afterQuery", async () => {
    const identity = await models.identity.create({
      email: "adam@keel.xyz",
    });

    const person = await models.person.create({
      title: "adam",
      sex: Sex.Male,
    });

    const deletedId = await actions
      .withIdentity(identity)
      .deletePersonAfterQuery({ id: person.id });

    expect(deletedId).toEqual(person.id);

    // test the afterQuery hook behaviour worked
    const log = await models.log.findMany();

    expect(log[0].msg).toEqual(`deleted person ${person.id}`);
  });

  test("get.beforeQuery", async () => {
    const identity = await models.identity.create({
      email: "adam@keel.xyz",
    });

    const person = await models.person.create({
      title: "adam",
      sex: Sex.Male,
    });

    // the beforeQuery adds a constraint that also adds gender = Female to the query
    await expect(
      actions.withIdentity(identity).getPersonBeforeQuery({
        id: person.id,
      })
    ).toHaveError({
      message: "no result",
    });
  });

  test("get.afterQuery", async () => {
    const identity = await models.identity.create({
      email: "adam@keel.xyz",
    });

    const person = await models.person.create({
      title: "adam",
      sex: Sex.Male,
    });
    await actions.withIdentity(identity).getPersonAfterQuery({
      id: person.id,
    });

    // test the afterQuery hook behaviour worked
    const log = await models.log.findMany();

    expect(log[0].msg).toEqual(`Fetched ${person.id}`);
  });

  test("list.beforeQuery", async () => {
    const identity = await models.identity.create({
      email: "adam@keel.xyz",
    });

    const person1 = await models.person.create({
      title: "adam",
      sex: Sex.Male,
    });

    const person2 = await models.person.create({
      title: "alice",
      sex: Sex.Female,
    });

    // the beforeQuery adds a constraint that also adds sex = Female to the query
    const data = await actions.withIdentity(identity).listPeopleBeforeQuery();

    expect(data.results.length).toEqual(1);
    expect(data.results.map((r) => r.id)).not.toContain(person1.id);
    expect(data.results.map((r) => r.id)).toContain(person2.id);
  });

  test("list.afterQuery", async () => {
    const identity = await models.identity.create({
      email: "adam@keel.xyz",
    });

    await models.person.create({
      title: "adam",
      sex: Sex.Male,
    });
    await actions.withIdentity(identity).listPeopleAfterQuery();

    // test the afterQuery hook behaviour worked
    const log = await models.log.findMany();

    expect(log[0].msg).toEqual(`List results: 1`);
  });

  test("update.beforeQuery", async () => {
    const identity = await models.identity.create({
      email: "adam@keel.xyz",
    });

    const person = await models.person.create({
      title: "alice",
      sex: Sex.Female,
    });

    // the beforeQuery tries to look up by the ID xxx instead
    // which will fail
    await expect(
      actions.withIdentity(identity).updatePersonWithBeforeQuery({
        where: {
          id: person.id,
        },
        values: {
          sex: Sex.Male,
          title: person.title,
        },
      })
    ).rejects.toThrowError("record not found");
  });

  test("update.afterQuery", async () => {
    const identity = await models.identity.create({
      email: "adam@keel.xyz",
    });

    const person = await models.person.create({
      title: "adam",
      sex: Sex.Male,
    });

    const data = await actions
      .withIdentity(identity)
      .updatePersonWithAfterQuery({
        where: {
          id: person.id,
        },
        values: {
          title: "bob",
          sex: Sex.Female,
        },
      });

    // afterQuery will prefix the updated title of the person with "not "
    expect(data.title).toEqual(`not bob`);
  });
});
