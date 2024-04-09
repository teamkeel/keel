import { models, actions, jobs, resetDatabase } from "@teamkeel/testing";
import {
  useDatabase,
  Wedding,
  Identity,
  WeddingInvitee,
  InviteStatus,
} from "@teamkeel/sdk";
import { test, expect, beforeEach } from "vitest";
import { sql } from "kysely";

beforeEach(resetDatabase);

interface Audit<T> {
  id: string;
  tableName: string;
  op: string;
  data: T;
  identityId: string | null;
  traceId: string | null;
  createdAt: Date;
}

test("create action - audit table populated", async () => {
  const wedding = await actions.createWedding({
    name: "Mary & Bob",
  });
  expect(wedding).not.toBeNull();

  const logs = await sql<Audit<Wedding>>`SELECT * FROM keel_audit`.execute(
    useDatabase()
  );
  expect(logs.rows.length).toEqual(1);
  const audit = logs.rows.at(0)!;

  expect(audit.id).toHaveLength(27);
  expect(audit.tableName).toEqual("wedding");
  expect(audit.op).toEqual("insert");
  expect(audit.identityId).toBeNull();
  expect(audit.createdAt).not.toBeNull();
  expect(audit.data.id).toEqual(wedding.id);
  expect(audit.data.name).toEqual(wedding.name);
  expect(audit.data.venue).toBeNull();
  expect(new Date(audit.data.createdAt).toISOString()).toEqual(
    new Date(wedding.createdAt).toISOString()
  );
  expect(new Date(audit.data.updatedAt).toISOString()).toEqual(
    new Date(wedding.updatedAt).toISOString()
  );
});

test("update action - audit table populated", async () => {
  const { id } = await actions.createWedding({
    name: "Mary & Bob",
  });
  expect(id).not.toBeNull();

  const wedding = await actions.updateWedding({
    where: { id: id },
    values: { name: "Mary and Bob" },
  });
  expect(wedding).not.toBeNull();

  const logs = await sql<
    Audit<Wedding>
  >`SELECT * FROM keel_audit WHERE op = 'update'`.execute(useDatabase());
  expect(logs.rows.length).toEqual(1);
  const audit = logs.rows.at(0)!;

  // Audit table columns
  expect(audit.id).toHaveLength(27);
  expect(audit.tableName).toEqual("wedding");
  expect(audit.op).toEqual("update");
  expect(audit.identityId).toBeNull();
  expect(audit.createdAt).not.toBeNull();

  // Data column
  expect(audit.data.id).toEqual(wedding.id);
  expect(audit.data.name).toEqual(wedding.name);
  expect(audit.data.venue).toBeNull();
  expect(new Date(audit.data.createdAt).toISOString()).toEqual(
    new Date(wedding.createdAt).toISOString()
  );
  expect(new Date(audit.data.updatedAt).toISOString()).toEqual(
    new Date(wedding.updatedAt).toISOString()
  );
});

test("delete action - audit table populated", async () => {
  const wedding = await actions.createWedding({
    name: "Mary & Bob",
  });
  expect(wedding).not.toBeNull();

  const id = await actions.deleteWedding({ id: wedding.id });
  expect(id).toEqual(wedding.id);

  const logs = await sql<
    Audit<Wedding>
  >`SELECT * FROM keel_audit WHERE op = 'delete'`.execute(useDatabase());
  expect(logs.rows.length).toEqual(1);
  const audit = logs.rows.at(0)!;

  // Audit table columns
  expect(audit.id).toHaveLength(27);
  expect(audit.tableName).toEqual("wedding");
  expect(audit.op).toEqual("delete");
  expect(audit.identityId).toBeNull();
  expect(audit.createdAt).not.toBeNull();

  // Data column
  expect(audit.data.id).toEqual(wedding.id);
  expect(audit.data.name).toEqual(wedding.name);
  expect(audit.data.venue).toBeNull();
  expect(new Date(audit.data.createdAt).toISOString()).toEqual(
    new Date(wedding.createdAt).toISOString()
  );
  expect(new Date(audit.data.updatedAt).toISOString()).toEqual(
    new Date(wedding.updatedAt).toISOString()
  );
});

test("create action with identity - audit table populated", async () => {
  const identity = await models.identity.create({
    email: "keelson@keel.xyz",
    issuer: "https://keel.so",
  });

  const wedding = await actions.withIdentity(identity).createWedding({
    name: "Mary & Bob",
  });
  expect(wedding).not.toBeNull();

  const logs = await sql<
    Audit<Wedding>
  >`SELECT * FROM keel_audit WHERE table_name = 'wedding'`.execute(
    useDatabase()
  );
  expect(logs.rows.length).toEqual(1);
  const audit = logs.rows.at(0)!;

  expect(audit.identityId).toEqual(identity.id);
  expect(audit.data.id).toEqual(wedding.id);
});

test("update action with identity - audit table populated", async () => {
  const identity = await models.identity.create({
    email: "keelson@keel.xyz",
    issuer: "https://keel.so",
  });

  const { id } = await actions.createWedding({
    name: "Mary & Bob",
  });
  expect(id).not.toBeNull();

  const wedding = await actions.withIdentity(identity).updateWedding({
    where: { id: id },
    values: { name: "Mary and Bob" },
  });
  expect(wedding).not.toBeNull();

  const logs = await sql<
    Audit<Wedding>
  >`SELECT * FROM keel_audit WHERE op = 'update'`.execute(useDatabase());
  expect(logs.rows.length).toEqual(1);
  const audit = logs.rows.at(0)!;

  expect(audit.identityId).toEqual(identity.id);
  expect(audit.data.id).toEqual(wedding.id);
});

test("delete action with identity - audit table populated", async () => {
  const identity = await models.identity.create({
    email: "keelson@keel.xyz",
    issuer: "https://keel.so",
  });

  const wedding = await actions.createWedding({
    name: "Mary & Bob",
  });
  expect(wedding).not.toBeNull();

  const id = await actions
    .withIdentity(identity)
    .deleteWedding({ id: wedding.id });
  expect(id).toEqual(wedding.id);

  const logs = await sql<
    Audit<Wedding>
  >`SELECT * FROM keel_audit WHERE op = 'delete'`.execute(useDatabase());
  expect(logs.rows.length).toEqual(1);
  const audit = logs.rows.at(0)!;

  expect(audit.identityId).toEqual(identity.id);
  expect(audit.data.id).toEqual(wedding.id);
});

test("nested create action - audit table populated", async () => {
  const identity = await models.identity.create({
    email: "mary@keel.xyz",
    issuer: "https://keel.so",
  });

  const wedding = await actions.withIdentity(identity).createWeddingWithGuests({
    name: "Mary & Bob",
    guests: [{ firstName: "Weave" }, { firstName: "Keelson" }],
  });
  expect(wedding).not.toBeNull();

  const logs = await sql<
    Audit<Wedding>
  >`SELECT * FROM keel_audit WHERE table_name = 'wedding'`.execute(
    useDatabase()
  );
  expect(logs.rows.length).toEqual(1);
  const weddingAudit = logs.rows.at(0)!;

  expect(weddingAudit.id).toHaveLength(27);
  expect(weddingAudit.tableName).toEqual("wedding");
  expect(weddingAudit.op).toEqual("insert");
  expect(weddingAudit.identityId).toEqual(identity.id);
  expect(weddingAudit.createdAt).not.toBeNull();

  expect(weddingAudit.data.id).toEqual(wedding.id);
  expect(weddingAudit.data.name).toEqual(wedding.name);
  expect(new Date(weddingAudit.data.createdAt).toISOString()).toEqual(
    new Date(wedding.createdAt).toISOString()
  );
  expect(new Date(weddingAudit.data.updatedAt).toISOString()).toEqual(
    new Date(wedding.updatedAt).toISOString()
  );

  const inviteelLogs = await sql<
    Audit<WeddingInvitee>
  >`SELECT * FROM keel_audit WHERE table_name = 'wedding_invitee'`.execute(
    useDatabase()
  );
  expect(inviteelLogs.rows.length).toEqual(2);

  const keelson = (
    await models.weddingInvitee.findMany({ where: { firstName: "Keelson" } })
  )[0];
  const keelsonLog = inviteelLogs.rows.at(0)!;

  expect(keelsonLog.id).not.toBeNull();
  expect(keelsonLog.tableName).toEqual("wedding_invitee");
  expect(keelsonLog.op).toEqual("insert");
  expect(keelsonLog.identityId).toEqual(identity.id);
  expect(keelsonLog.createdAt).not.toBeNull();

  expect(keelsonLog.data.id).toEqual(keelson.id);
  expect(keelsonLog.data.firstName).toEqual(keelson.firstName);
  expect(keelsonLog.data.weddingId).toEqual(wedding.id);
  expect(new Date(keelsonLog.data.createdAt).toISOString()).toEqual(
    new Date(keelson.createdAt).toISOString()
  );
  expect(new Date(keelsonLog.data.updatedAt).toISOString()).toEqual(
    new Date(keelson.updatedAt).toISOString()
  );

  const weave = (
    await models.weddingInvitee.findMany({ where: { firstName: "Weave" } })
  )[0];
  const weaveLog = inviteelLogs.rows.at(1)!;

  expect(weaveLog.id).not.toBeNull();
  expect(weaveLog.tableName).toEqual("wedding_invitee");
  expect(weaveLog.op).toEqual("insert");
  expect(weaveLog.identityId).toEqual(identity.id);
  expect(weaveLog.createdAt).not.toBeNull();

  expect(weaveLog.data.id).toEqual(weave.id);
  expect(weaveLog.data.firstName).toEqual(weave.firstName);
  expect(weaveLog.data.weddingId).toEqual(wedding.id);
  expect(new Date(weaveLog.data.createdAt).toISOString()).toEqual(
    new Date(weave.createdAt).toISOString()
  );
  expect(new Date(weaveLog.data.updatedAt).toISOString()).toEqual(
    new Date(weave.updatedAt).toISOString()
  );
});

test("built-in actions with multiple identities - audit table populated", async () => {
  const keelson = await models.identity.create({
    email: "keelson@keel.xyz",
    issuer: "https://keel.so",
  });
  const weave = await models.identity.create({
    email: "weave@keel.xyz",
    issuer: "https://keel.so",
  });

  const wedding = await actions.withIdentity(keelson).createWedding({
    name: "Mary & Bob",
  });
  expect(wedding).not.toBeNull();

  const updated = await actions.updateWedding({
    where: { id: wedding.id },
    values: { name: "Mary and Bob" },
  });
  expect(updated.id).toEqual(wedding.id);

  const deleted = await actions
    .withIdentity(weave)
    .deleteWedding({ id: wedding.id });
  expect(deleted).toEqual(wedding.id);

  const logs = await sql<
    Audit<Wedding>
  >`SELECT * FROM keel_audit WHERE table_name = 'wedding'`.execute(
    useDatabase()
  );
  expect(logs.rows.length).toEqual(3);

  const insertAudit = logs.rows.at(0)!;
  expect(insertAudit.identityId).toEqual(keelson.id);
  expect(insertAudit.data.id).toEqual(wedding.id);

  const updateAudit = logs.rows.at(1)!;
  expect(updateAudit.identityId).toBeNull();
  expect(updateAudit.data.id).toEqual(wedding.id);

  const deleteAudit = logs.rows.at(2)!;
  expect(deleteAudit.identityId).toEqual(weave.id);
  expect(deleteAudit.data.id).toEqual(wedding.id);
});

test("hook function - audit table populated", async () => {
  const identity = await models.identity.create({
    email: "keelson@keel.xyz",
    issuer: "https://keel.so",
  });

  const guest = await actions.withIdentity(identity).inviteGuest({
    firstName: "Keelson",
    isFamily: true,
  });

  const logs = await sql<
    Audit<WeddingInvitee>
  >`SELECT * FROM keel_audit WHERE table_name = 'wedding_invitee'`.execute(
    useDatabase()
  );
  expect(logs.rows.length).toEqual(2);

  const insertAudit = logs.rows.at(0)!;

  expect(insertAudit.id).toHaveLength(27);
  expect(insertAudit.tableName).toEqual("wedding_invitee");
  expect(insertAudit.op).toEqual("insert");
  expect(insertAudit.identityId).toEqual(identity.id);
  expect(insertAudit.createdAt).not.toBeNull();

  expect(insertAudit.data.id).toEqual(guest.id);
  expect(insertAudit.data.firstName).toEqual(guest.firstName);
  expect(insertAudit.data.isFamily).toBeTruthy();
  expect(insertAudit.data.status).toEqual(InviteStatus.Pending);
  expect(insertAudit.data.weddingId).toBeNull();
  expect(new Date(insertAudit.data.createdAt).toISOString()).toEqual(
    new Date(guest.createdAt).toISOString()
  );
  expect(new Date(insertAudit.data.updatedAt).toISOString()).toEqual(
    new Date(guest.updatedAt).toISOString()
  );

  const updateAudit = logs.rows.at(1)!;

  expect(updateAudit.id).toHaveLength(27);
  expect(updateAudit.tableName).toEqual("wedding_invitee");
  expect(updateAudit.op).toEqual("update");
  expect(updateAudit.identityId).toEqual(identity.id);
  expect(updateAudit.createdAt).not.toBeNull();

  expect(updateAudit.data.id).toEqual(guest.id);
  expect(updateAudit.data.firstName).toEqual(guest.firstName);
  expect(updateAudit.data.isFamily).toBeTruthy();
  expect(updateAudit.data.status).toEqual(InviteStatus.Accepted);
  expect(updateAudit.data.weddingId).toBeNull();
  expect(new Date(updateAudit.data.createdAt).toISOString()).toEqual(
    new Date(guest.createdAt).toISOString()
  );
  expect(new Date(updateAudit.data.updatedAt).toISOString()).toEqual(
    new Date(guest.updatedAt).toISOString()
  );
});

test("write function with identity - audit table populated", async () => {
  const identity = await models.identity.create({
    email: "keelson@keel.xyz",
    issuer: "https://keel.so",
  });

  const wedding = await actions.createWedding({
    name: "Mary & Bob",
  });
  expect(wedding).not.toBeNull();

  await actions.withIdentity(identity).inviteMany({
    names: ["Keelson", "Weave"],
    weddingId: wedding.id,
  });

  const logs = await sql<
    Audit<WeddingInvitee>
  >`SELECT * FROM keel_audit WHERE table_name = 'wedding_invitee'`.execute(
    useDatabase()
  );
  expect(logs.rows.length).toEqual(2);

  const insertKeelson = logs.rows.at(0)!;

  expect(insertKeelson.id).not.toBeNull();
  expect(insertKeelson.tableName).toEqual("wedding_invitee");
  expect(insertKeelson.op).toEqual("insert");
  expect(insertKeelson.identityId).toEqual(identity.id);
  expect(insertKeelson.createdAt).not.toBeNull();

  expect(insertKeelson.data.firstName).toEqual("Keelson");
  expect(insertKeelson.data.isFamily).toBeFalsy();
  expect(insertKeelson.data.status).toEqual(InviteStatus.Pending);
  expect(insertKeelson.data.weddingId).toEqual(wedding.id);

  const insertWeave = logs.rows.at(1)!;

  expect(insertWeave.id).not.toBeNull();
  expect(insertWeave.tableName).toEqual("wedding_invitee");
  expect(insertWeave.op).toEqual("insert");
  expect(insertWeave.identityId).toEqual(identity.id);
  expect(insertWeave.createdAt).not.toBeNull();

  expect(insertWeave.data.firstName).toEqual("Weave");
  expect(insertWeave.data.isFamily).toBeFalsy();
  expect(insertWeave.data.status).toEqual(InviteStatus.Pending);
  expect(insertWeave.data.weddingId).toEqual(wedding.id);
});

test("write function with error and rollback - model and audit table empty", async () => {
  const identity = await models.identity.create({
    email: "keelson@keel.xyz",
    issuer: "https://keel.so",
  });

  const wedding = await actions.createWedding({
    name: "Mary & Bob",
  });
  expect(wedding).not.toBeNull();

  await expect(
    actions.withIdentity(identity).inviteMany({
      names: ["Keelson", "Weave", "Prisma"],
      weddingId: wedding.id,
    })
  ).toHaveError({ message: "prisma is not invited!" });

  const guests = await models.weddingInvitee.findMany();
  expect(guests).toHaveLength(0);

  const logs = await sql<
    Audit<WeddingInvitee>
  >`SELECT * FROM keel_audit WHERE table_name = 'wedding_invitee'`.execute(
    useDatabase()
  );
  expect(logs.rows.length).toEqual(0);
});

test("job function with identity - audit table populated", async () => {
  const identity = await models.identity.create({
    email: "keelson@keel.xyz",
    issuer: "https://keel.so",
  });

  const wedding = await actions.createWedding({
    name: "Mary & Bob",
  });
  expect(wedding).not.toBeNull();

  const keelson = await models.weddingInvitee.create({
    firstName: "Keelson",
    status: InviteStatus.Accepted,
    weddingId: wedding.id,
  });
  const keeler = await models.weddingInvitee.create({
    firstName: "Keeler",
    status: InviteStatus.Declined,
    weddingId: wedding.id,
  });
  const weaveton = await models.weddingInvitee.create({
    firstName: "Weaveton",
    status: InviteStatus.Pending,
    weddingId: wedding.id,
  });

  await jobs.withIdentity(identity).updateHeadCount({ weddingId: wedding.id });

  const inviteesAudits = await sql<
    Audit<WeddingInvitee>
  >`SELECT * FROM keel_audit where table_name = 'wedding_invitee'`.execute(
    useDatabase()
  );
  expect(inviteesAudits.rows.length).toEqual(4);

  const keelsonAudit = inviteesAudits.rows.at(0)!;
  expect(keelsonAudit.id).toHaveLength(27);
  expect(keelsonAudit.tableName).toEqual("wedding_invitee");
  expect(keelsonAudit.op).toEqual("insert");
  expect(keelsonAudit.identityId).toBeNull();
  expect(keelsonAudit.data.id).toEqual(keelson.id);

  const keelerAudit = inviteesAudits.rows.at(1)!;
  expect(keelsonAudit.id).toHaveLength(27);
  expect(keelerAudit.tableName).toEqual("wedding_invitee");
  expect(keelerAudit.op).toEqual("insert");
  expect(keelerAudit.identityId).toBeNull();
  expect(keelerAudit.data.id).toEqual(keeler.id);

  const weavetonAudit = inviteesAudits.rows.at(2)!;
  expect(weavetonAudit.id).toHaveLength(27);
  expect(weavetonAudit.tableName).toEqual("wedding_invitee");
  expect(weavetonAudit.op).toEqual("insert");
  expect(weavetonAudit.identityId).toBeNull();
  expect(weavetonAudit.data.id).toEqual(weaveton.id);

  const keelerDeleteAudit = inviteesAudits.rows.at(3)!;
  expect(keelerDeleteAudit.id).toHaveLength(27);
  expect(keelerDeleteAudit.tableName).toEqual("wedding_invitee");
  expect(keelerDeleteAudit.op).toEqual("delete");
  expect(keelerDeleteAudit.identityId).toEqual(identity.id);
  expect(keelerDeleteAudit.data.id).toEqual(keeler.id);
  expect(keelerDeleteAudit.data.firstName).toEqual(keeler.firstName);
  expect(keelerDeleteAudit.data.status).toEqual(keeler.status);
  expect(keelerDeleteAudit.data.isFamily).toEqual(keeler.isFamily);

  const weddingAudits = await sql<
    Audit<Wedding>
  >`SELECT * FROM keel_audit where table_name = 'wedding'`.execute(
    useDatabase()
  );
  expect(weddingAudits.rows.length).toEqual(2);

  const weddingAudit = weddingAudits.rows.at(0)!;
  expect(weddingAudit.id).toHaveLength(27);
  expect(weddingAudit.tableName).toEqual("wedding");
  expect(weddingAudit.op).toEqual("insert");
  expect(weddingAudit.identityId).toBeNull();
  expect(weddingAudit.data.id).toEqual(wedding.id);
  expect(weddingAudit.data.name).toEqual(wedding.name);
  expect(weddingAudit.data.headcount).toEqual(0);

  const weddingUpdateAudit = weddingAudits.rows.at(1)!;
  expect(weddingUpdateAudit.id).toHaveLength(27);
  expect(weddingUpdateAudit.tableName).toEqual("wedding");
  expect(weddingUpdateAudit.op).toEqual("update");
  expect(weddingUpdateAudit.identityId).toEqual(identity.id);
  expect(weddingUpdateAudit.data.id).toEqual(wedding.id);
  expect(weddingUpdateAudit.data.name).toEqual(wedding.name);
  expect(weddingUpdateAudit.data.headcount).toEqual(1);
});

test("job function with error and no rollback - audit table is not rolled back", async () => {
  const identity = await models.identity.create({
    email: "keelson@keel.xyz",
    issuer: "https://keel.so",
  });

  const wedding = await actions.createWedding({
    name: "Mary & Bob",
  });
  expect(wedding).not.toBeNull();

  const keelson = await models.weddingInvitee.create({
    firstName: "Keelson",
    status: InviteStatus.Accepted,
    weddingId: wedding.id,
  });
  const keeler = await models.weddingInvitee.create({
    firstName: "Keeler",
    status: InviteStatus.Declined,
    weddingId: wedding.id,
  });
  const prisma = await models.weddingInvitee.create({
    firstName: "Prisma",
    status: InviteStatus.Pending,
    weddingId: wedding.id,
  });

  await expect(
    jobs.withIdentity(identity).updateHeadCount({ weddingId: wedding.id })
  ).toHaveError({ message: "prisma is not invited!" });

  const inviteesAudits = await sql<
    Audit<WeddingInvitee>
  >`SELECT * FROM keel_audit where table_name = 'wedding_invitee'`.execute(
    useDatabase()
  );
  expect(inviteesAudits.rows.length).toEqual(4);

  const keelsonAudit = inviteesAudits.rows.at(0)!;
  expect(keelsonAudit.id).toHaveLength(27);
  expect(keelsonAudit.tableName).toEqual("wedding_invitee");
  expect(keelsonAudit.op).toEqual("insert");
  expect(keelsonAudit.identityId).toBeNull();
  expect(keelsonAudit.data.id).toEqual(keelson.id);

  const keelerAudit = inviteesAudits.rows.at(1)!;
  expect(keelsonAudit.id).toHaveLength(27);
  expect(keelerAudit.tableName).toEqual("wedding_invitee");
  expect(keelerAudit.op).toEqual("insert");
  expect(keelerAudit.identityId).toBeNull();
  expect(keelerAudit.data.id).toEqual(keeler.id);

  const weavetonAudit = inviteesAudits.rows.at(2)!;
  expect(weavetonAudit.id).toHaveLength(27);
  expect(weavetonAudit.tableName).toEqual("wedding_invitee");
  expect(weavetonAudit.op).toEqual("insert");
  expect(weavetonAudit.identityId).toBeNull();
  expect(weavetonAudit.data.id).toEqual(prisma.id);

  const keelerDeleteAudit = inviteesAudits.rows.at(3)!;
  expect(keelerDeleteAudit.id).toHaveLength(27);
  expect(keelerDeleteAudit.tableName).toEqual("wedding_invitee");
  expect(keelerDeleteAudit.op).toEqual("delete");
  expect(keelerDeleteAudit.identityId).toEqual(identity.id);
  expect(keelerDeleteAudit.data.id).toEqual(keeler.id);
  expect(keelerDeleteAudit.data.firstName).toEqual(keeler.firstName);
  expect(keelerDeleteAudit.data.status).toEqual(keeler.status);
  expect(keelerDeleteAudit.data.isFamily).toEqual(keeler.isFamily);

  const weddingAudits = await sql<
    Audit<Wedding>
  >`SELECT * FROM keel_audit where table_name = 'wedding'`.execute(
    useDatabase()
  );
  expect(weddingAudits.rows.length).toEqual(2);

  const weddingAudit = weddingAudits.rows.at(0)!;
  expect(weddingAudit.id).toHaveLength(27);
  expect(weddingAudit.tableName).toEqual("wedding");
  expect(weddingAudit.op).toEqual("insert");
  expect(weddingAudit.identityId).toBeNull();
  expect(weddingAudit.data.id).toEqual(wedding.id);
  expect(weddingAudit.data.name).toEqual(wedding.name);
  expect(weddingAudit.data.headcount).toEqual(0);

  const weddingUpdateAudit = weddingAudits.rows.at(1)!;
  expect(weddingUpdateAudit.id).toHaveLength(27);
  expect(weddingUpdateAudit.tableName).toEqual("wedding");
  expect(weddingUpdateAudit.op).toEqual("update");
  expect(weddingUpdateAudit.identityId).toEqual(identity.id);
  expect(weddingUpdateAudit.data.id).toEqual(wedding.id);
  expect(weddingUpdateAudit.data.name).toEqual(wedding.name);
  expect(weddingUpdateAudit.data.headcount).toEqual(1);
});

test("job function using kysely with identity - audit table populated", async () => {
  const identity = await models.identity.create({
    email: "keelson@keel.xyz",
    issuer: "https://keel.so",
  });

  const wedding = await actions.createWedding({
    name: "Mary & Bob",
  });
  expect(wedding).not.toBeNull();

  const keelson = await models.weddingInvitee.create({
    firstName: "Keelson",
    status: InviteStatus.Accepted,
    weddingId: wedding.id,
  });
  const keeler = await models.weddingInvitee.create({
    firstName: "Keeler",
    status: InviteStatus.Declined,
    weddingId: wedding.id,
  });
  const weaveton = await models.weddingInvitee.create({
    firstName: "Weaveton",
    status: InviteStatus.Pending,
    weddingId: wedding.id,
  });

  await jobs
    .withIdentity(identity)
    .updateHeadCountWithKysely({ weddingId: wedding.id });

  const inviteesAudits = await sql<
    Audit<WeddingInvitee>
  >`SELECT * FROM keel_audit where table_name = 'wedding_invitee'`.execute(
    useDatabase()
  );
  expect(inviteesAudits.rows.length).toEqual(4);

  const keelsonAudit = inviteesAudits.rows.at(0)!;
  expect(keelsonAudit.id).toHaveLength(27);
  expect(keelsonAudit.tableName).toEqual("wedding_invitee");
  expect(keelsonAudit.op).toEqual("insert");
  expect(keelsonAudit.identityId).toBeNull();
  expect(keelsonAudit.data.id).toEqual(keelson.id);

  const keelerAudit = inviteesAudits.rows.at(1)!;
  expect(keelsonAudit.id).toHaveLength(27);
  expect(keelerAudit.tableName).toEqual("wedding_invitee");
  expect(keelerAudit.op).toEqual("insert");
  expect(keelerAudit.identityId).toBeNull();
  expect(keelerAudit.data.id).toEqual(keeler.id);

  const weavetonAudit = inviteesAudits.rows.at(2)!;
  expect(weavetonAudit.id).toHaveLength(27);
  expect(weavetonAudit.tableName).toEqual("wedding_invitee");
  expect(weavetonAudit.op).toEqual("insert");
  expect(weavetonAudit.identityId).toBeNull();
  expect(weavetonAudit.data.id).toEqual(weaveton.id);

  const keelerDeleteAudit = inviteesAudits.rows.at(3)!;
  expect(keelerDeleteAudit.id).toHaveLength(27);
  expect(keelerDeleteAudit.tableName).toEqual("wedding_invitee");
  expect(keelerDeleteAudit.op).toEqual("delete");
  expect(keelerDeleteAudit.identityId).toEqual(identity.id);
  expect(keelerDeleteAudit.data.id).toEqual(keeler.id);
  expect(keelerDeleteAudit.data.firstName).toEqual(keeler.firstName);
  expect(keelerDeleteAudit.data.status).toEqual(keeler.status);
  expect(keelerDeleteAudit.data.isFamily).toEqual(keeler.isFamily);

  const weddingAudits = await sql<
    Audit<Wedding>
  >`SELECT * FROM keel_audit where table_name = 'wedding'`.execute(
    useDatabase()
  );
  expect(weddingAudits.rows.length).toEqual(2);

  const weddingAudit = weddingAudits.rows.at(0)!;
  expect(weddingAudit.id).toHaveLength(27);
  expect(weddingAudit.tableName).toEqual("wedding");
  expect(weddingAudit.op).toEqual("insert");
  expect(weddingAudit.identityId).toBeNull();
  expect(weddingAudit.data.id).toEqual(wedding.id);
  expect(weddingAudit.data.name).toEqual(wedding.name);
  expect(weddingAudit.data.headcount).toEqual(0);

  const weddingUpdateAudit = weddingAudits.rows.at(1)!;
  expect(weddingUpdateAudit.id).toHaveLength(27);
  expect(weddingUpdateAudit.tableName).toEqual("wedding");
  expect(weddingUpdateAudit.op).toEqual("update");
  expect(weddingUpdateAudit.identityId).toEqual(identity.id);
  expect(weddingUpdateAudit.data.id).toEqual(wedding.id);
  expect(weddingUpdateAudit.data.name).toEqual(wedding.name);
  expect(weddingUpdateAudit.data.headcount).toEqual(1);
});

test("identity model - audit table populated", async () => {
  await models.identity.create({
    email: "user@keel.xyz",
    issuer: "https://keel.so",
  });

  const identity = await models.identity.findOne({
    email: "user@keel.xyz",
    issuer: "https://keel.so",
  });

  expect(identity).not.toBeNull();

  const logs = await sql<Audit<Identity>>`SELECT * FROM keel_audit`.execute(
    useDatabase()
  );
  expect(logs.rows.length).toEqual(1);
  const audit = logs.rows.at(0)!;

  expect(audit.id).toHaveLength(27);
  expect(audit.tableName).toEqual("identity");
  expect(audit.op).toEqual("insert");
  expect(audit.identityId).toBeNull();
  expect(audit.createdAt).not.toBeNull();

  expect(audit.data.id).toEqual(identity!.id);
  expect(audit.data.email).toEqual(identity!.email);
  expect(audit.data.password).toEqual(identity!.password);
  expect(audit.data.emailVerified).toEqual(identity!.emailVerified);
  expect(audit.data.externalId).toEqual(identity!.externalId);
  expect(audit.data.issuer).toEqual(identity!.issuer);
  expect(new Date(audit.data.createdAt).toISOString()).toEqual(
    new Date(identity!.createdAt).toISOString()
  );
  expect(new Date(audit.data.updatedAt).toISOString()).toEqual(
    new Date(identity!.updatedAt).toISOString()
  );

  const updated = await models.identity.update(
    { id: identity!.id },
    { email: "dave@keel.xyz" }
  );
  const updateLogs = await sql<
    Audit<Identity>
  >`SELECT * FROM keel_audit WHERE table_name = 'identity' AND op = 'update'`.execute(
    useDatabase()
  );
  expect(updateLogs.rows.length).toEqual(1);
  const updateLog = updateLogs.rows.at(0)!;
  expect(updated.id).toEqual(updateLog!.data.id);

  const deleted = await models.identity.delete({ id: identity!.id });
  const deletedLogs = await sql<
    Audit<Identity>
  >`SELECT * FROM keel_audit WHERE table_name = 'identity' AND op = 'delete'`.execute(
    useDatabase()
  );
  expect(deletedLogs.rows.length).toEqual(1);
  const deleteLog = deletedLogs.rows.at(0)!;
  expect(deleted).toEqual(deleteLog!.data.id);
});

test("model API in integration tests - audit table is populated without identity or trace IDs", async () => {
  const created = await models.wedding.create({ name: "Mary & Bob" });
  const updated = await models.wedding.update(
    { id: created.id },
    { name: "Mary and Bob" }
  );
  const deleted = await models.wedding.delete({ id: created.id });

  const logs = await sql<Audit<Wedding>>`SELECT * FROM keel_audit`.execute(
    useDatabase()
  );
  expect(logs.rows.length).toEqual(3);

  const insertLog = logs.rows.at(0)!;
  expect(insertLog.op).toEqual("insert");
  expect(insertLog.identityId).toBeNull();
  expect(insertLog.traceId).toBeNull();
  expect(insertLog.data.id).toEqual(created.id);
  expect(insertLog.data.name).toEqual(created.name);

  const updateLog = logs.rows.at(1)!;
  expect(updateLog.op).toEqual("update");
  expect(insertLog.identityId).toBeNull();
  expect(insertLog.traceId).toBeNull();
  expect(updateLog.data.id).toEqual(updated.id);
  expect(updateLog.data.name).toEqual(updated.name);

  const deleteLog = logs.rows.at(2)!;
  expect(deleteLog.op).toEqual("delete");
  expect(insertLog.identityId).toBeNull();
  expect(insertLog.traceId).toBeNull();
  expect(deleteLog.data.id).toEqual(deleted);
  expect(deleteLog.data.name).toEqual(updated.name);
});
