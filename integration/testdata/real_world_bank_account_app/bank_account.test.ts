import { actions, models } from "@teamkeel/testing";
import { test, expect, beforeAll } from "vitest";

let bigEntity;
let bigAccount;
let identityJohn, userJohn;
let identityBob, userBob;

let smallEntity;
let smallAccount;
let identityHenry, userHenry;

let identityDave, adminDave;
let identityBenoit, adminBenoit;
let identityTom, adminTom;

beforeAll(async () => {
  // Big Org with 2 users
  bigEntity = await models.entity.create({ name: "Big Org" });
  bigAccount = await models.bankAccount.create({
    balance: 0,
    entityId: bigEntity.id,
  });

  identityJohn = await models.identity.create({ email: "john@bigorg.so" });
  userJohn = await models.entityUser.create({
    name: "John",
    entityId: bigEntity.id,
    identityId: identityJohn.id,
    canUpdate: true,
  });

  identityBob = await models.identity.create({ email: "bob@bigorg.so" });
  userBob = await models.entityUser.create({
    name: "Bob",
    entityId: bigEntity.id,
    identityId: identityBob.id,
    canUpdate: false,
  });

  // Small Org with 1 user
  smallEntity = await models.entity.create({ name: "Small Org" });
  smallAccount = await models.bankAccount.create({
    balance: 0,
    entityId: smallEntity.id,
  });

  identityHenry = await models.identity.create({ email: "henry@smallorg.so" });
  userHenry = await models.entityUser.create({
    name: "Henry",
    entityId: smallEntity.id,
    identityId: identityHenry.id,
    canUpdate: true,
  });

  // Administrator / Dave access to Big and Small Org
  identityDave = await models.identity.create({ email: "dave@bank.com" });
  adminDave = await models.administrator.create({
    identityId: identityDave.id,
  });
  await models.entityAccess.create({
    entityId: bigEntity.id,
    adminId: adminDave.id,
  });
  await models.entityAccess.create({
    entityId: smallEntity.id,
    adminId: adminDave.id,
  });

  // Administrator / Benoit access to Big Org
  identityBenoit = await models.identity.create({ email: "benoit@bank.com" });
  adminBenoit = await models.administrator.create({
    identityId: identityBenoit.id,
  });
  await models.entityAccess.create({
    entityId: bigEntity.id,
    adminId: adminBenoit.id,
  });

  // Administrator / No access to orgs
  identityTom = await models.identity.create({ email: "tom@bank.com" });
  adminTom = await models.administrator.create({
    identityId: identityTom.id,
  });
});

test("myUser returns correct user", async () => {
  const result = await actions.withIdentity(identityJohn).myUser();

  expect(result).not.toBeNull();
  expect(result?.id).toEqual(userJohn.id);
});

test("myUser returns null for identity which is not a user", async () => {
  const rouge = await models.identity.create({ email: "rouge@internet.com" });

  const result = await actions.withIdentity(rouge).myUser();
  expect(result).toBeNull();
});

test("myAccount returns correct entity's account", async () => {
  const result = await actions.withIdentity(identityJohn).myAccount();

  expect(result).not.toBeNull();
  expect(result?.id).toEqual(bigAccount.id);
});

test("updateMyAccount updates correct entity's account", async () => {
  const result = await actions
    .withIdentity(identityJohn)
    .updateMyAccount({ values: { alias: "Cheque Account" } });

  expect(result).not.toBeNull();
  expect(result?.id).toEqual(bigAccount.id);
  expect(result?.alias).toEqual("Cheque Account");

  const updatedAccount = await models.bankAccount.findOne({
    id: bigAccount.id,
  });

  expect(updatedAccount).not.toBeNull();
  expect(updatedAccount?.id).toEqual(bigAccount.id);
  expect(updatedAccount?.alias).toEqual("Cheque Account");
});

test("updateMyAccount returns authorization error because canUpdate is false for user", async () => {
  await expect(
    actions
      .withIdentity(identityBob)
      .updateMyAccount({ values: { alias: "Cheque Account" } })
  ).toHaveAuthorizationError();
});

test("getAccount authenticates against admins without access", async () => {
  const noResult = await actions
    .withIdentity(identityJohn)
    .getAccount({ id: "unknown" });
  expect(noResult).toBeNull();

  await expect(
    actions.getAccount({ id: bigAccount.id })
  ).toHaveAuthorizationError();

  await expect(
    actions.withIdentity(identityJohn).getAccount({ id: bigAccount.id })
  ).toHaveAuthorizationError();

  await expect(
    actions.withIdentity(identityTom).getAccount({ id: bigAccount.id })
  ).toHaveAuthorizationError();

  const result = await actions
    .withIdentity(identityDave)
    .getAccount({ id: bigAccount.id });
  expect(result).not.toBeNull();
  expect(result?.id).toEqual(bigAccount.id);
});

test("myFellowUsers retrieves the users for an entity", async () => {
  const results = await actions.withIdentity(identityJohn).myFellowUsers();

  expect(results.pageInfo.count).toEqual(1);
  expect(results.results[0].name).toEqual("Bob");
});

test("myFellowUsers retrieves no users for an admin identity", async () => {
  const results = await actions.withIdentity(identityDave).myFellowUsers();

  expect(results.pageInfo.count).toEqual(0);
});

test("entityUsers retrieves the users for an entity", async () => {
  const result1 = await actions.withIdentity(identityDave).entityUsers({
    where: {
      entity: {
        id: {
          equals: bigEntity.id,
        },
      },
    },
  });

  expect(result1.pageInfo.totalCount).toEqual(2);
  expect(result1.results[0].name).toEqual("Bob");
  expect(result1.results[1].name).toEqual("John");

  const result2 = await actions.withIdentity(identityDave).entityUsers({
    where: {
      entity: {
        id: {
          equals: smallEntity.id,
        },
      },
    },
  });

  expect(result2.pageInfo.totalCount).toEqual(1);
  expect(result2.results[0].name).toEqual("Henry");

  const result3 = await actions.withIdentity(identityBenoit).entityUsers({
    where: {
      entity: {
        id: {
          equals: bigEntity.id,
        },
      },
    },
  });

  expect(result3.pageInfo.totalCount).toEqual(2);
  expect(result3.results[0].name).toEqual("Bob");
  expect(result3.results[1].name).toEqual("John");
});

test("entityUsers returns authorization error as identity is not admin of entity", async () => {
  await expect(
    actions.entityUsers({
      where: {
        entity: {
          id: {
            equals: bigEntity.id,
          },
        },
      },
    })
  ).toHaveAuthorizationError();

  await expect(
    actions.withIdentity(identityBenoit).entityUsers({
      where: {
        entity: {
          id: {
            equals: smallEntity.id,
          },
        },
      },
    })
  ).toHaveAuthorizationError();

  await expect(
    actions.withIdentity(identityTom).entityUsers({
      where: {
        entity: {
          id: {
            equals: smallEntity.id,
          },
        },
      },
    })
  ).toHaveAuthorizationError();

  await expect(
    actions.withIdentity(identityTom).entityUsers({
      where: {
        entity: {
          id: {
            equals: bigEntity.id,
          },
        },
      },
    })
  ).toHaveAuthorizationError();
});

test("bankAccountUsers retrieves the users by bank account", async () => {
  const result = await actions.withIdentity(identityDave).bankAccountUsers({
    where: {
      entity: {
        account: {
          id: {
            equals: bigAccount.id,
          },
        },
      },
    },
  });

  expect(result.pageInfo.totalCount).toEqual(2);
  expect(result.results[0].name).toEqual("Bob");
  expect(result.results[1].name).toEqual("John");

  const result2 = await actions.withIdentity(identityDave).bankAccountUsers({
    where: {
      entity: {
        account: {
          id: {
            equals: smallAccount.id,
          },
        },
      },
    },
  });

  expect(result2.pageInfo.totalCount).toEqual(1);
  expect(result2.results[0].name).toEqual("Henry");

  const result3 = await actions.withIdentity(identityBenoit).bankAccountUsers({
    where: {
      entity: {
        account: {
          id: {
            equals: bigAccount.id,
          },
        },
      },
    },
  });

  expect(result3.pageInfo.totalCount).toEqual(2);
  expect(result3.results[0].name).toEqual("Bob");
  expect(result3.results[1].name).toEqual("John");
});

test("bankAccountUsers returns authorization error as identity is not admin of entity", async () => {
  await expect(
    actions.withIdentity(identityBenoit).bankAccountUsers({
      where: {
        entity: {
          account: {
            id: {
              equals: smallAccount.id,
            },
          },
        },
      },
    })
  ).toHaveAuthorizationError();

  await expect(
    actions.bankAccountUsers({
      where: {
        entity: {
          account: {
            id: {
              equals: bigAccount.id,
            },
          },
        },
      },
    })
  ).toHaveAuthorizationError();

  await expect(
    actions.withIdentity(identityTom).bankAccountUsers({
      where: {
        entity: {
          account: {
            id: {
              equals: bigAccount.id,
            },
          },
        },
      },
    })
  ).toHaveAuthorizationError();

  await expect(
    actions.withIdentity(identityTom).bankAccountUsers({
      where: {
        entity: {
          account: {
            id: {
              equals: smallAccount.id,
            },
          },
        },
      },
    })
  ).toHaveAuthorizationError();
});
