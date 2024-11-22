import { actions, resetDatabase, models } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

// test("create - @set with backlinks", async () => {
//   const org = await models.organisation.create({
//     name: "Keel",
//     isActive: true,
//   });
//   const identity = await models.identity.create({ email: "keelson@keel.so" });
//   const user = await models.user.create({
//     name: "Keelson",
//     identityId: identity.id,
//     organisationId: org.id,
//   });
//   const record = await actions
//     .withIdentity(identity)
//     .createRecord({ name: "Tax Records" });

//   expect(record).not.toBeNull();
//   expect(record.ownerId).toEqual(user.id);
//   expect(record.organisationId).toEqual(org.id);
//   expect(record.isActive).toEqual(true);
// });

// test("create - @set with backlinks and 1:M nested create", async () => {
//   const org = await models.organisation.create({
//     name: "Keel",
//     isActive: true,
//   });
//   const identity = await models.identity.create({ email: "keelson@keel.so" });

//   const user = await models.user.create({
//     name: "Keelson",
//     identityId: identity.id,
//     organisationId: org.id,
//   });
//   const record = await actions.withIdentity(identity).createRecordWithChildren({
//     name: "Tax Records",
//     children: [{ name: "VAT" }, { name: "Income Tax" }, { name: "PAYE Tax" }],
//   });

//   expect(record).not.toBeNull();
//   expect(record.ownerId).toEqual(user.id);
//   expect(record.organisationId).toEqual(org.id);
//   expect(record.isActive).toEqual(true);
//   expect(record.parentId).toBeNull();

//   const children = await models.record.findMany({
//     where: { parentId: record.id },
//   });
//   expect(children).toHaveLength(3);

//   expect(children[0].isActive).toEqual(true);
//   expect(children[0].ownerId).toEqual(user.id);
//   expect(children[0].organisationId).toEqual(org.id);

//   expect(children[1].isActive).toEqual(true);
//   expect(children[1].ownerId).toEqual(user.id);
//   expect(children[1].organisationId).toEqual(org.id);

//   expect(children[2].isActive).toEqual(true);
//   expect(children[2].ownerId).toEqual(user.id);
//   expect(children[2].organisationId).toEqual(org.id);
// });

// test("create - @set with backlinks and M:1 nested create", async () => {
//   const org = await models.organisation.create({
//     name: "Keel",
//     isActive: true,
//   });
//   const identity = await models.identity.create({ email: "keelson@keel.so" });
//   const user = await models.user.create({
//     name: "Keelson",
//     identityId: identity.id,
//     organisationId: org.id,
//   });
//   const record = await actions.withIdentity(identity).createRecordWithParent({
//     name: "Tax Records",
//     parent: { name: "Operations" },
//   });

//   expect(record).not.toBeNull();
//   expect(record.ownerId).toEqual(user.id);
//   expect(record.organisationId).toEqual(org.id);
//   expect(record.isActive).toEqual(true);
//   expect(record.parentId).not.toBeNull();

//   const parent = await models.record.findMany({
//     where: { id: { equals: record.parentId } },
//   });

//   expect(parent).toHaveLength(1);
//   expect(parent[0].ownerId).toEqual(user.id);
//   expect(parent[0].organisationId).toEqual(org.id);
//   expect(parent[0].isActive).toEqual(true);
//   expect(parent[0].parentId).toBeNull();
// });

// test("create - @set with backlinks and no authenticated identity", async () => {
//   const org = await models.organisation.create({
//     name: "Keel",
//     isActive: true,
//   });
//   const identity = await models.identity.create({ email: "keelson@keel.so" });
//   const user = await models.user.create({
//     name: "Keelson",
//     identityId: identity.id,
//     organisationId: org.id,
//   });

//   await expect(actions.createRecord({ name: "Tax Records" })).toHaveError({
//     code: "ERR_INVALID_INPUT",
//     message: "field 'ownerId' cannot be null",
//   });
// });

// test("create - @set with backlinks and no user backlink", async () => {
//   const identity = await models.identity.create({ email: "keelson@keel.so" });

//   await expect(
//     actions.withIdentity(identity).createRecord({ name: "Tax Records" })
//   ).toHaveError({
//     code: "ERR_INVALID_INPUT",
//     message: "field 'ownerId' cannot be null",
//   });
// });

// test("update - @set with backlinks", async () => {
//   const orgKeel = await models.organisation.create({
//     name: "Keel",
//     isActive: true,
//   });
//   const identityKeelson = await models.identity.create({
//     email: "keelson@keel.so",
//   });
//   const userKeelson = await models.user.create({
//     name: "Keelson",
//     identityId: identityKeelson.id,
//     organisationId: orgKeel.id,
//   });

//   const record = await actions.withIdentity(identityKeelson).createRecord({
//     name: "Tax Records",
//   });

//   const orgWeave = await models.organisation.create({
//     name: "Weave",
//     isActive: false,
//   });
//   const identityWeaveton = await models.identity.create({
//     email: "weaveton@keel.so",
//   });
//   const userWeaveton = await models.user.create({
//     name: "Weaveton",
//     identityId: identityWeaveton.id,
//     organisationId: orgWeave.id,
//   });

//   const updatedRecord = await actions
//     .withIdentity(identityWeaveton)
//     .updateRecordOwner({
//       where: { id: record.id },
//     });

//   expect(updatedRecord).not.toBeNull();
//   expect(updatedRecord.ownerId).toEqual(userWeaveton.id);
//   expect(updatedRecord.organisationId).toEqual(orgWeave.id);
//   expect(updatedRecord.isActive).toEqual(false);
//   expect(updatedRecord.parentId).toBeNull();
// });

// test("update - @set with backlinks and no authenticated identity", async () => {
//   const org = await models.organisation.create({
//     name: "Keel",
//     isActive: true,
//   });
//   const identity = await models.identity.create({ email: "keelson@keel.so" });
//   const user = await models.user.create({
//     name: "Keelson",
//     identityId: identity.id,
//     organisationId: org.id,
//   });

//   const record = await actions.withIdentity(identity).createRecord({
//     name: "Tax Records",
//   });

//   await expect(
//     actions.updateRecordOwner({
//       where: { id: record.id },
//     })
//   ).toHaveError({
//     code: "ERR_INVALID_INPUT",
//     message: "field 'ownerId' cannot be null",
//   });
// });

// test("update - @set with backlinks and no user backlink", async () => {
//   const org = await models.organisation.create({
//     name: "Keel",
//     isActive: true,
//   });
//   const identity = await models.identity.create({ email: "keelson1@keel.so" });
//   const user = await models.user.create({
//     name: "Keelson",
//     identityId: identity.id,
//     organisationId: org.id,
//   });

//   const record = await actions.withIdentity(identity).createRecord({
//     name: "Tax Records",
//   });

//   const identity2 = await models.identity.create({ email: "keelson2@keel.so" });

//   await expect(
//     actions.withIdentity(identity2).updateRecordOwner({
//       where: { id: record.id },
//     })
//   ).toHaveError({
//     code: "ERR_INVALID_INPUT",
//     message: "field 'ownerId' cannot be null",
//   });
// });

test("create - @set with identity fields", async () => {
  const { id } = await models.identity.create({
    email: "user@keel.xyz",
    issuer: "https://keel.so",
  });

  const identity = await models.identity.update(
    { id: id },
    { externalId: "extId" }
  );

  const org = await models.organisation.create({
    name: "Keel",
    isActive: true,
  });
  const user = await models.user.create({
    name: "Keelson",
    identityId: identity!.id,
    organisationId: org.id,
  });

  const extension = await actions
    .withIdentity(identity!)
    .createExt({ n: "Keelson" });

  expect(extension.name).toEqual("Keelson");
  expect(extension.identity1Id).toEqual(identity?.id);
  expect(extension.identity2Id).toEqual(identity?.id);
  expect(extension.user1Id).toEqual(user.id);
  expect(extension.user2Id).toEqual(user.id);
  expect(extension.email).toEqual(identity?.email);
  expect(extension.isVerified).toEqual(identity?.emailVerified);
  expect(extension.issuer).toEqual(identity?.issuer);
  // https://linear.app/keel/issue/KE-1192/datetime-precision-loss
  //expect(extension.signedUpAt).toEqual(identity?.createdAt);
  expect(extension.externalId).toEqual(identity?.externalId);
});

// test("update - @set with identity fields", async () => {
//   const { id: identityId } = await models.identity.create({
//     email: "user@keel.xyz",
//     issuer: "https://keel.so",
//   });

//   const identity = await models.identity.update(
//     { id: identityId },
//     { externalId: "extId" }
//   );

//   const org = await models.organisation.create({
//     name: "Keel",
//     isActive: true,
//   });
//   const user = await models.user.create({
//     name: "Keelson",
//     identityId: identity!.id,
//     organisationId: org.id,
//   });

//   const { id } = await actions
//     .withIdentity(identity!)
//     .createExt({ n: "Keelson" });

//   const extension = await actions
//     .withIdentity(identity!)
//     .updateExt({ where: { id: id }, values: { n: "Keelson" } });

//   expect(extension.name).toEqual("Keelson");
//   expect(extension.identity1Id).toEqual(identity?.id);
//   expect(extension.identity2Id).toEqual(identity?.id);
//   expect(extension.user1Id).toEqual(user.id);
//   expect(extension.user2Id).toEqual(user.id);
//   expect(extension.email).toEqual(identity?.email);
//   expect(extension.isVerified).toEqual(identity?.emailVerified);
//   expect(extension.issuer).toEqual(identity?.issuer);
//   // https://linear.app/keel/issue/KE-1192/datetime-precision-loss
//   //expect(extension.signedUpAt).toEqual(identity?.createdAt);
//   expect(extension.externalId).toEqual(identity?.externalId);
// });
