import { test, expect, beforeEach } from "vitest";
const { ModelAPI } = require("./ModelAPI");
const { PROTO_ACTION_TYPES } = require("./consts");
const { sql } = require("kysely");
const { useDatabase, withDatabase } = require("./database");
const KSUID = require("ksuid");
const TraceParent = require("traceparent");
const { withAuditContext } = require("./auditing");

let personAPI;
const db = useDatabase();

beforeEach(async () => {
  await sql`
 
  DROP TABLE IF EXISTS post;
  DROP TABLE IF EXISTS person;
  DROP TABLE IF EXISTS author;

  CREATE TABLE person(
      id               text PRIMARY KEY,
      name             text UNIQUE
  );

  CREATE OR REPLACE FUNCTION set_identity_id(id VARCHAR)
  RETURNS TEXT AS $$
  BEGIN
      RETURN set_config('audit.identity_id', id, true);
  END
  $$ LANGUAGE plpgsql;

  CREATE OR REPLACE FUNCTION set_trace_id(id VARCHAR)
  RETURNS TEXT AS $$
  BEGIN
      RETURN set_config('audit.trace_id', id, true);
  END
  $$ LANGUAGE plpgsql;
  `.execute(db);

  personAPI = new ModelAPI("person", undefined, {});
});

async function identityIdFromConfigParam(database, nonLocal = true) {
  const result =
    await sql`SELECT NULLIF(current_setting('audit.identity_id', ${sql.literal(
      nonLocal
    )}), '') AS id`.execute(database);
  return result.rows[0].id;
}

async function traceIdFromConfigParam(database, nonLocal = true) {
  const result =
    await sql`SELECT NULLIF(current_setting('audit.trace_id', ${sql.literal(
      nonLocal
    )}), '') AS id`.execute(database);
  return result.rows[0].id;
}

test("auditing - capturing identity id in transaction", async () => {
  const request = {
    meta: {
      identity: { id: KSUID.randomSync().string },
    },
  };

  const identityId = request.meta.identity.id;

  const row = await withDatabase(
    db,
    PROTO_ACTION_TYPES.CREATE, // CREATE will ensure a transaction is opened
    async ({ transaction }) => {
      const row = withAuditContext(request, async () => {
        return await personAPI.create({
          id: KSUID.randomSync().string,
          name: "James",
        });
      });

      expect(await identityIdFromConfigParam(transaction)).toEqual(identityId);
      expect(await identityIdFromConfigParam(db)).toBeNull();

      return row;
    }
  );

  expect(row.name).toEqual("James");
  expect(KSUID.parse(row.id).string).toEqual(row.id);

  expect(await identityIdFromConfigParam(db)).toBeNull();
  expect(await identityIdFromConfigParam(db, false)).toBeNull();
});

test("auditing - capturing tracing in transaction", async () => {
  const request = {
    meta: {
      tracing: {
        traceparent: "00-80e1afed08e019fc1110464cfa66635c-7a085853722dc6d2-01",
      },
    },
  };

  const traceId = TraceParent.fromString(
    request.meta.tracing.traceparent
  ).traceId;

  const row = await withDatabase(
    db,
    PROTO_ACTION_TYPES.CREATE, // CREATE will ensure a transaction is opened
    async ({ transaction }) => {
      const row = withAuditContext(request, async () => {
        return await personAPI.create({
          id: KSUID.randomSync().string,
          name: "Jim",
        });
      });

      expect(await traceIdFromConfigParam(transaction)).toEqual(traceId);
      expect(await traceIdFromConfigParam(db)).toBeNull();

      return row;
    }
  );

  expect(KSUID.parse(row.id).string).toEqual(row.id);
  expect(row.name).toEqual("Jim");

  expect(await traceIdFromConfigParam(db)).toBeNull();
  expect(await traceIdFromConfigParam(db, false)).toBeNull();
});

test("auditing - capturing identity id without transaction", async () => {
  const request = {
    meta: {
      identity: { id: KSUID.randomSync().string },
    },
  };

  const row = await withDatabase(
    db,
    PROTO_ACTION_TYPES.GET, // GET will _not_ open a transaction
    async ({ sDb }) => {
      const row = withAuditContext(request, async () => {
        return await personAPI.create({
          id: KSUID.randomSync().string,
          name: "James",
        });
      });

      expect(await identityIdFromConfigParam(sDb)).toBeNull();
      expect(await identityIdFromConfigParam(db)).toBeNull();

      return row;
    }
  );

  expect(row.name).toEqual("James");
  expect(KSUID.parse(row.id).string).toEqual(row.id);

  expect(await identityIdFromConfigParam(db)).toBeNull();
  expect(await identityIdFromConfigParam(db, false)).toBeNull();
});

test("auditing - capturing tracing without transaction", async () => {
  const request = {
    meta: {
      tracing: {
        traceparent: "00-80e1afed08e019fc1110464cfa66635c-7a085853722dc6d2-01",
      },
    },
  };

  const row = await withDatabase(
    db,
    PROTO_ACTION_TYPES.GET, // GET will _not_ open a transaction
    async ({ sDb }) => {
      const row = withAuditContext(request, async () => {
        return await personAPI.create({
          id: KSUID.randomSync().string,
          name: "Jim",
        });
      });

      expect(await traceIdFromConfigParam(sDb)).toBeNull();
      expect(await traceIdFromConfigParam(db)).toBeNull();

      return row;
    }
  );

  expect(KSUID.parse(row.id).string).toEqual(row.id);
  expect(row.name).toEqual("Jim");

  expect(await traceIdFromConfigParam(db)).toBeNull();
  expect(await traceIdFromConfigParam(db, false)).toBeNull();
});

test("auditing - no audit context", async () => {
  const row = await withDatabase(
    db,
    PROTO_ACTION_TYPES.CREATE,
    async ({ transaction }) => {
      const row = withAuditContext({}, async () => {
        return await personAPI.create({
          id: KSUID.randomSync().string,
          name: "Jake",
        });
      });

      expect(await identityIdFromConfigParam(transaction)).toBeNull();
      expect(await identityIdFromConfigParam(db)).toBeNull();
      expect(await traceIdFromConfigParam(transaction)).toBeNull();
      expect(await traceIdFromConfigParam(db)).toBeNull();
      return row;
    }
  );

  expect(KSUID.parse(row.id).string).toEqual(row.id);
  expect(row.name).toEqual("Jake");
});

test("auditing - ModelAPI.create", async () => {
  const request = {
    meta: {
      identity: { id: KSUID.randomSync().string },
      tracing: {
        traceparent: "00-80e1afed08e019fc1110464cfa66635c-7a085853722dc6d2-01",
      },
    },
  };

  const identityId = request.meta.identity.id;
  const traceId = TraceParent.fromString(
    request.meta.tracing.traceparent
  ).traceId;

  const row = await withDatabase(
    db,
    PROTO_ACTION_TYPES.CREATE,
    async ({ transaction }) => {
      const row = withAuditContext(request, async () => {
        return await personAPI.create({
          id: KSUID.randomSync().string,
          name: "Jake",
        });
      });

      expect(await identityIdFromConfigParam(transaction)).toEqual(identityId);
      expect(await traceIdFromConfigParam(transaction)).toEqual(traceId);

      return row;
    }
  );

  expect(row.name).toEqual("Jake");
});

test("auditing - ModelAPI.update", async () => {
  const request = {
    meta: {
      identity: { id: KSUID.randomSync().string },
      tracing: {
        traceparent: "00-80e1afed08e019fc1110464cfa66635c-7a085853722dc6d2-01",
      },
    },
  };

  const identityId = request.meta.identity.id;
  const traceId = TraceParent.fromString(
    request.meta.tracing.traceparent
  ).traceId;

  const created = await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Jake",
  });

  const row = await withDatabase(
    db,
    PROTO_ACTION_TYPES.CREATE,
    async ({ transaction }) => {
      const row = withAuditContext(request, async () => {
        return await personAPI.update({ id: created.id }, { name: "Jim" });
      });

      expect(await identityIdFromConfigParam(transaction)).toEqual(identityId);
      expect(await traceIdFromConfigParam(transaction)).toEqual(traceId);

      return row;
    }
  );

  expect(row.name).toEqual("Jim");
});

test("auditing - ModelAPI.delete", async () => {
  const request = {
    meta: {
      identity: { id: KSUID.randomSync().string },
      tracing: {
        traceparent: "00-80e1afed08e019fc1110464cfa66635c-7a085853722dc6d2-01",
      },
    },
  };

  const identityId = request.meta.identity.id;
  const traceId = TraceParent.fromString(
    request.meta.tracing.traceparent
  ).traceId;

  const created = await personAPI.create({
    id: KSUID.randomSync().string,
    name: "Jake",
  });

  const row = await withDatabase(
    db,
    PROTO_ACTION_TYPES.CREATE,
    async ({ transaction }) => {
      const row = withAuditContext(request, async () => {
        return await personAPI.delete({ id: created.id });
      });

      expect(await identityIdFromConfigParam(transaction)).toEqual(identityId);
      expect(await traceIdFromConfigParam(transaction)).toEqual(traceId);

      return row;
    }
  );

  expect(row).toEqual(created.id);
});

test("auditing - identity id and trace id fields dropped from result", async () => {
  const request = {
    meta: {
      identity: { id: KSUID.randomSync().string },
      tracing: {
        traceparent: "00-80e1afed08e019fc1110464cfa66635c-7a085853722dc6d2-01",
      },
    },
  };

  const identityId = request.meta.identity.id;
  const traceId = TraceParent.fromString(
    request.meta.tracing.traceparent
  ).traceId;

  const row = await withDatabase(
    db,
    PROTO_ACTION_TYPES.CREATE,
    async ({ transaction }) => {
      const row = withAuditContext(request, async () => {
        return await personAPI.create({
          id: KSUID.randomSync().string,
          name: "Jake",
        });
      });

      expect(await identityIdFromConfigParam(transaction)).toEqual(identityId);
      expect(await traceIdFromConfigParam(transaction)).toEqual(traceId);

      return row;
    }
  );

  expect(row.name).toEqual("Jake");
  expect(row.keelIdentityId).toBeUndefined();
  expect(row.keelTraceId).toBeUndefined();
  expect(Object.keys(row).length).toEqual(2);
});
