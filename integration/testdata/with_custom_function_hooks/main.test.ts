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


  test("update action - updatedAt set", async () => {
    const identity = await models.identity.create({
      email: "adam@keel.xyz",
    });
    const name = "Alice";

    const person = await models.person.create({
      sex: Sex.Female,
      title: name,
    });

    expect(person.updatedAt).not.toBeNull();
    expect(person.updatedAt).toEqual(person.createdAt);

    await delay(100);

     const record = await actions
      .withIdentity(identity)
      .updatePersonWithBeforeWrite({
        where: { id: person.id },
        values: {
          title: "Alice",
          sex: Sex.Female,
        },
      });
      console.log(person);
      console.log(record);

      const person2 = await models.person.findOne({id: person.id});
  console.log(person2);


    expect(record.updatedAt.valueOf()).toBeGreaterThanOrEqual(person.createdAt.valueOf() + 100);
    expect(record.updatedAt.valueOf()).toBeLessThan(person.createdAt.valueOf() + 1000);
  });

  function delay(ms: number) {
    return new Promise( resolve => setTimeout(resolve, ms) );
  }
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

  // this is a special case because the result returned from the beforeWrite hook is passed onto
  // the beforeQuery hook so therefore we want to test that the beforeQuery hooks receives the mutated
  // version of the inputs from the beforeWrite hook.
  test("update.beforeWrite + beforeQuery combination", async () => {
    const identity = await models.identity.create({
      email: "adam@keel.xyz",
    });

    const person = await models.person.create({
      sex: Sex.Male,
      title: "adam",
    });

    const result = await actions
      .withIdentity(identity)
      .updatePersonWithBeforeWriteAndBeforeQuery({
        where: { id: person.id },
        values: {
          sex: Sex.Male,
          title: "bob",
        },
      });

    // the afterQuery hook calls .repeat(2) on the result of the beforeWrite hook which also calls .repeat(2)
    // so we expect the result returned from the action to be 4 bobs
    expect(result.title).toEqual("bobbobbobbob");

    const updatedRecord = await models.person.findOne({ id: person.id });

    // and in the database we only expect bob to be repeated once
    expect(updatedRecord?.title).toEqual("bobbob");
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
