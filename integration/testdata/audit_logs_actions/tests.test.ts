import { models, actions, resetDatabase } from "@teamkeel/testing"
import { useDatabase, Wedding, Identity, WeddingInvitee , InviteStatus } from "@teamkeel/sdk"
import { test, expect, beforeEach } from "vitest"
import { sql } from "kysely"

beforeEach( resetDatabase);

interface Audit<T> {
    id: String;
    tableName: String;
    op: String;
    data: T;
    identityId: String | null;
    traceId: String | null;
    createdAt: Date;
}

test("create action - audit table is populated correctly", async () => {
    const wedding = await actions.createWedding({
      name: "Mary & Bob",
    });
    expect(wedding).not.toBeNull();

    const logs = await sql<Audit<Wedding>>`SELECT * FROM keel_audit`.execute(useDatabase());
    expect(logs.rows.length).toEqual(1);
    const audit = logs.rows.at(0)!;

    // Audit table columns
    expect(audit.id).not.toBeNull();
    expect(audit.tableName).toEqual("wedding");
    expect(audit.op).toEqual("insert");
    expect(audit.identityId).toBeNull();
    expect(audit.traceId).toBeNull();
    expect(audit.createdAt).not.toBeNull();

    // Data column
    expect(audit.data.id).toEqual(wedding.id);
    expect(audit.data.name).toEqual(wedding.name);
    expect(new Date(audit.data.createdAt).toISOString()).toEqual(new Date(wedding.createdAt).toISOString());
    expect(new Date(audit.data.updatedAt).toISOString()).toEqual(new Date(wedding.updatedAt).toISOString());
  });

  test("update action - audit table is populated correctly", async () => {
    const { id } = await actions.createWedding({
      name: "Mary & Bob",
    });
    expect(id).not.toBeNull();

    const wedding = await actions.updateWedding({
        where: { id: id },
        values: { name: "Mary and Bob" }
    });
      expect(wedding).not.toBeNull();

    const logs = await sql<Audit<Wedding>>`SELECT * FROM keel_audit WHERE op = 'update'`.execute(useDatabase());
    expect(logs.rows.length).toEqual(1);
    const audit = logs.rows.at(0)!;

    // Audit table columns
    expect(audit.id).not.toBeNull();
    expect(audit.tableName).toEqual("wedding");
    expect(audit.op).toEqual("update");
    expect(audit.identityId).toBeNull();
    expect(audit.traceId).toBeNull();
    expect(audit.createdAt).not.toBeNull();

    // Data column
    expect(audit.data.id).toEqual(wedding.id);
    expect(audit.data.name).toEqual(wedding.name);
    expect(new Date(audit.data.createdAt).toISOString()).toEqual(new Date(wedding.createdAt).toISOString());
    expect(new Date(audit.data.updatedAt).toISOString()).toEqual(new Date(wedding.updatedAt).toISOString());
  });


  test("delete action - audit table is populated correctly", async () => {
    const wedding = await actions.createWedding({
      name: "Mary & Bob",
    });
    expect(wedding).not.toBeNull();

    const id = await actions.deleteWedding({id: wedding.id });
    expect(id).toEqual(wedding.id);

    const logs = await sql<Audit<Wedding>>`SELECT * FROM keel_audit WHERE op = 'delete'`.execute(useDatabase());
    expect(logs.rows.length).toEqual(1);
    const audit = logs.rows.at(0)!;

    // Audit table columns
    expect(audit.id).not.toBeNull();
    expect(audit.tableName).toEqual("wedding");
    expect(audit.op).toEqual("delete");
    expect(audit.identityId).toBeNull();
    expect(audit.traceId).toBeNull();
    expect(audit.createdAt).not.toBeNull();

    // Data column
    const data = audit.data;
    expect(data.id).toEqual(wedding.id);
    expect(data.name).toEqual(wedding.name);
    expect(new Date(audit.data.createdAt).toISOString()).toEqual(new Date(wedding.createdAt).toISOString());
    expect(new Date(audit.data.updatedAt).toISOString()).toEqual(new Date(wedding.updatedAt).toISOString());
  });

  test("create action - with identity", async () => {
    const identity = await models.identity.create({ email: "keelson@keel.xyz" })

    const wedding = await actions.withIdentity(identity).createWedding({
      name: "Mary & Bob",
    });
    expect(wedding).not.toBeNull();

    const logs = await sql<Audit<Wedding>>`SELECT * FROM keel_audit WHERE table_name = 'wedding'`.execute(useDatabase());
    expect(logs.rows.length).toEqual(1);
    const audit = logs.rows.at(0)!;

    expect(audit.identityId).toEqual(identity.id);
    expect(audit.traceId).toBeNull();
    expect(audit.data.id).toEqual(wedding.id);
  });


  test("update action - with identity", async () => {
    const identity = await models.identity.create({ email: "keelson@keel.xyz" })

    const { id } = await actions.createWedding({
      name: "Mary & Bob",
    });
    expect(id).not.toBeNull();

    const wedding = await actions.withIdentity(identity).updateWedding({
        where: { id: id },
        values: { name: "Mary and Bob" }
    });
      expect(wedding).not.toBeNull();

    const logs = await sql<Audit<Wedding>>`SELECT * FROM keel_audit WHERE op = 'update'`.execute(useDatabase());
    expect(logs.rows.length).toEqual(1);
    const audit = logs.rows.at(0)!;

    expect(audit.identityId).toEqual(identity.id);
    expect(audit.traceId).toBeNull();
    expect(audit.data.id).toEqual(wedding.id);
  });



  test("delete action - with identity", async () => {
    const identity = await models.identity.create({ email: "keelson@keel.xyz" })

    const wedding = await actions.createWedding({
      name: "Mary & Bob",
    });
    expect(wedding).not.toBeNull();

    const id = await actions.withIdentity(identity).deleteWedding({id: wedding.id });
    expect(id).toEqual(wedding.id);

    const logs = await sql<Audit<Wedding>>`SELECT * FROM keel_audit WHERE op = 'delete'`.execute(useDatabase());
    expect(logs.rows.length).toEqual(1);
    const audit = logs.rows.at(0)!;

    expect(audit.identityId).toEqual(identity.id);
    expect(audit.traceId).toBeNull();
    expect(audit.data.id).toEqual(wedding.id);
  });


  test("nested create action - audit table is populated correctly", async () => {
    const identity = await models.identity.create({ email: "mary@keel.xyz" })

    const wedding = await actions.withIdentity(identity).createWeddingWithGuests({
      name: "Mary & Bob",
      guests: [
        { firstName: "Weave"},
        { firstName: "Keelson" }
      ]
    });
    expect(wedding).not.toBeNull();

    const logs = await sql<Audit<Wedding>>`SELECT * FROM keel_audit WHERE table_name = 'wedding'`.execute(useDatabase());
    expect(logs.rows.length).toEqual(1);
    const weddingAudit = logs.rows.at(0)!;

    // Audit table columns
    expect(weddingAudit.id).not.toBeNull();
    expect(weddingAudit.tableName).toEqual("wedding");
    expect(weddingAudit.op).toEqual("insert");
    expect(weddingAudit.identityId).toEqual(identity.id);
    expect(weddingAudit.traceId).toBeNull();
    expect(weddingAudit.createdAt).not.toBeNull();

    // Data column
    expect(weddingAudit.data.id).toEqual(wedding.id);
    expect(weddingAudit.data.name).toEqual(wedding.name);
    expect(new Date(weddingAudit.data.createdAt).toISOString()).toEqual(new Date(wedding.createdAt).toISOString());
    expect(new Date(weddingAudit.data.updatedAt).toISOString()).toEqual(new Date(wedding.updatedAt).toISOString());

    const inviteelLogs = await sql<Audit<WeddingInvitee>>`SELECT * FROM keel_audit WHERE table_name = 'wedding_invitee'`.execute(useDatabase());
    expect(inviteelLogs.rows.length).toEqual(2);

    const keelson = (await models.weddingInvitee.findMany({ where: { firstName: "Keelson"}}))[0];
    const keelsonLog = inviteelLogs.rows.at(0)!;

    // Audit table columns
    expect(keelsonLog.id).not.toBeNull();
    expect(keelsonLog.tableName).toEqual("wedding_invitee");
    expect(keelsonLog.op).toEqual("insert");
    expect(keelsonLog.identityId).toEqual(identity.id);
    expect(keelsonLog.traceId).toBeNull();
    expect(keelsonLog.createdAt).not.toBeNull();

    // Data column
    expect(keelsonLog.data.id).toEqual(keelson.id);
    expect(keelsonLog.data.firstName).toEqual(keelson.firstName);
    expect(keelsonLog.data.weddingId).toEqual(wedding.id);
    expect(new Date(keelsonLog.data.createdAt).toISOString()).toEqual(new Date(keelson.createdAt).toISOString());
    expect(new Date(keelsonLog.data.updatedAt).toISOString()).toEqual(new Date(keelson.updatedAt).toISOString());

    const weave = (await models.weddingInvitee.findMany({ where: { firstName: "Weave"}}))[0];
    const weaveLog = inviteelLogs.rows.at(1)!;

    // Audit table columns
    expect(weaveLog.id).not.toBeNull();
    expect(weaveLog.tableName).toEqual("wedding_invitee");
    expect(weaveLog.op).toEqual("insert");
    expect(weaveLog.identityId).toEqual(identity.id);
    expect(weaveLog.traceId).toBeNull();
    expect(weaveLog.createdAt).not.toBeNull();

    // Data column
    expect(weaveLog.data.id).toEqual(weave.id);
    expect(weaveLog.data.firstName).toEqual(weave.firstName);
    expect(weaveLog.data.weddingId).toEqual(wedding.id);
    expect(new Date(weaveLog.data.createdAt).toISOString()).toEqual(new Date(weave.createdAt).toISOString());
    expect(new Date(weaveLog.data.updatedAt).toISOString()).toEqual(new Date(weave.updatedAt).toISOString());
  });

  test("built-in actions - multiple identities", async () => {
    const keelson = await models.identity.create({ email: "keelson@keel.xyz" });
    const weave = await models.identity.create({ email: "weave@keel.xyz" });

    const wedding = await actions.withIdentity(keelson).createWedding({
      name: "Mary & Bob",
    });
    expect(wedding).not.toBeNull();

    const updated = await actions.updateWedding({
        where: { id: wedding.id },
        values: { name: "Mary and Bob" }
    });
      expect(updated.id).toEqual(wedding.id);

      const deleted = await actions.withIdentity(weave).deleteWedding({id: wedding.id });
    expect(deleted).toEqual(wedding.id);


    const logs = await sql<Audit<Wedding>>`SELECT * FROM keel_audit WHERE table_name = 'wedding'`.execute(useDatabase());
    expect(logs.rows.length).toEqual(3);

    const insertAudit = logs.rows.at(0)!;
    expect(insertAudit.identityId).toEqual(keelson.id);
    expect(insertAudit.traceId).toBeNull();
    expect(insertAudit.data.id).toEqual(wedding.id);

    const updateAudit = logs.rows.at(1)!;
    expect(updateAudit.identityId).toBeNull();
    expect(updateAudit.traceId).toBeNull();
    expect(updateAudit.data.id).toEqual(wedding.id);

    const deleteAudit = logs.rows.at(2)!;
    expect(deleteAudit.identityId).toEqual(weave.id);
    expect(deleteAudit.traceId).toBeNull();
    expect(deleteAudit.data.id).toEqual(wedding.id);
  });

  test("hook function - audit table is populated correctly", async () => {
    const identity = await models.identity.create({ email: "keelson@keel.xyz" })

    const guest = await actions.withIdentity(identity).inviteGuest({
      firstName: "Keelson",
      isFamily: true
    });

    const logs = await sql<Audit<WeddingInvitee>>`SELECT * FROM keel_audit WHERE table_name = 'wedding_invitee'`.execute(useDatabase());
    expect(logs.rows.length).toEqual(2);
    
    const insertAudit = logs.rows.at(0)!;

    // Audit table columns
    expect(insertAudit.id).not.toBeNull();
    expect(insertAudit.tableName).toEqual("wedding_invitee");
    expect(insertAudit.op).toEqual("insert");
    expect(insertAudit.identityId).toEqual(identity.id);
    expect(insertAudit.traceId).toBeNull();
    expect(insertAudit.createdAt).not.toBeNull();

    // Data column
    expect(insertAudit.data.id).toEqual(guest.id);
    expect(insertAudit.data.firstName).toEqual(guest.firstName);
    expect(insertAudit.data.isFamily).toBeTruthy();
    expect(insertAudit.data.status).toEqual(InviteStatus.Pending);
    expect(insertAudit.data.weddingId).toBeNull();
    expect(new Date(insertAudit.data.createdAt).toISOString()).toEqual(new Date(guest.createdAt).toISOString());
    expect(new Date(insertAudit.data.updatedAt).toISOString()).toEqual(new Date(guest.updatedAt).toISOString());

    const updateAudit = logs.rows.at(1)!;

    // Audit table columns
    expect(updateAudit.id).not.toBeNull();
    expect(updateAudit.tableName).toEqual("wedding_invitee");
    expect(updateAudit.op).toEqual("update");
    expect(updateAudit.identityId).toEqual(identity.id);
    expect(updateAudit.traceId).toBeNull();
    expect(updateAudit.createdAt).not.toBeNull();

    // Data column
    expect(updateAudit.data.id).toEqual(guest.id);
    expect(updateAudit.data.firstName).toEqual(guest.firstName);
    expect(updateAudit.data.isFamily).toBeTruthy();
    expect(updateAudit.data.status).toEqual(InviteStatus.Accepted);
    expect(updateAudit.data.weddingId).toBeNull();
    expect(new Date(updateAudit.data.createdAt).toISOString()).toEqual(new Date(guest.createdAt).toISOString());
    expect(new Date(updateAudit.data.updatedAt).toISOString()).toEqual(new Date(guest.updatedAt).toISOString());

  });

  test("write function - audit table is populated correctly", async () => {
    const identity = await models.identity.create({ email: "keelson@keel.xyz" })

    const wedding = await actions.createWedding({
      name: "Mary & Bob",
    });
    expect(wedding).not.toBeNull();


    const result = await actions.withIdentity(identity).inviteMany({
      names: ["Keelson", "Weave"],
      weddingId: wedding.id
    });

    const logs = await sql<Audit<WeddingInvitee>>`SELECT * FROM keel_audit WHERE table_name = 'wedding_invitee'`.execute(useDatabase());
    expect(logs.rows.length).toEqual(2);
    
    const insertKeelson = logs.rows.at(0)!;

    // Audit table columns
    expect(insertKeelson.id).not.toBeNull();
    expect(insertKeelson.tableName).toEqual("wedding_invitee");
    expect(insertKeelson.op).toEqual("insert");
    expect(insertKeelson.identityId).toEqual(identity.id);
    expect(insertKeelson.traceId).toBeNull();
    expect(insertKeelson.createdAt).not.toBeNull();

    // Data column
    expect(insertKeelson.data.firstName).toEqual("Keelson");
    expect(insertKeelson.data.isFamily).toBeFalsy();
    expect(insertKeelson.data.status).toEqual(InviteStatus.Pending);
    expect(insertKeelson.data.weddingId).toEqual(wedding.id);

    const insertWeave = logs.rows.at(1)!;

    // Audit table columns
    expect(insertWeave.id).not.toBeNull();
    expect(insertWeave.tableName).toEqual("wedding_invitee");
    expect(insertWeave.op).toEqual("update");
    expect(insertWeave.identityId).toEqual(identity.id);
    expect(insertWeave.traceId).toBeNull();
    expect(insertWeave.createdAt).not.toBeNull();

    // Data column
    expect(insertWeave.data.firstName).toEqual("Weave");
    expect(insertWeave.data.isFamily).toBeFalsy();
    expect(insertWeave.data.status).toEqual(InviteStatus.Pending);
    expect(insertWeave.data.weddingId).toEqual(wedding.id);

  });

  test("identity model - audit table is populated correctly", async () => {
    const { identityCreated } = await actions.authenticate({
        createIfNotExists: true,
        emailPassword: {
          email: "user@keel.xyz",
          password: "1234",
        },
      });
      expect(identityCreated).toEqual(true);

      const identity = await models.identity.findOne({email: "user@keel.xyz"});
    expect(identity).not.toBeNull();

      const logs = await sql<Audit<Identity>>`SELECT * FROM keel_audit`.execute(useDatabase());
      expect(logs.rows.length).toEqual(1);
      const audit = logs.rows.at(0)!;

       // Audit table columns
    expect(audit.id).not.toBeNull();
    expect(audit.tableName).toEqual("identity");
    expect(audit.op).toEqual("insert");
    expect(audit.identityId).toBeNull();
    expect(audit.traceId).toBeNull();
    expect(audit.createdAt).not.toBeNull();

    // Data column
    expect(audit.data.id).toEqual(identity!.id);
    expect(audit.data.email).toEqual(identity!.email);
    expect(audit.data.password).toEqual(identity!.password);
    expect(audit.data.emailVerified).toEqual(identity!.emailVerified);
    expect(audit.data.externalId).toEqual(identity!.externalId);
    expect(audit.data.issuer).toEqual(identity!.issuer);
    expect(new Date(audit.data.createdAt).toISOString()).toEqual(new Date(identity!.createdAt).toISOString());
    expect(new Date(audit.data.updatedAt).toISOString()).toEqual(new Date(identity!.updatedAt).toISOString());

    const updated = await models.identity.update({id: identity!.id},  { email: "dave@keel.xyz"});
    const updateLogs = await sql<Audit<Identity>>`SELECT * FROM keel_audit WHERE table_name = 'identity' AND op = 'update'`.execute(useDatabase());
    expect(updateLogs.rows.length).toEqual(1);
    const updateLog = updateLogs.rows.at(0)!;
    expect(updated.id).toEqual(updateLog!.data.id);

    const deleted = await models.identity.delete({id: identity!.id});
    const deletedLogs = await sql<Audit<Identity>>`SELECT * FROM keel_audit WHERE table_name = 'identity' AND op = 'delete'`.execute(useDatabase());
    expect(deletedLogs.rows.length).toEqual(1);
    const deleteLog = deletedLogs.rows.at(0)!;
    expect(deleted).toEqual(deleteLog!.data.id);
  });