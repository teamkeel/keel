import {
  Permissions,
  PERMISSION_STATE,
  PermissionError,
  checkBuiltInPermissions,
} from "./permissions";
import { getDatabase } from "./database";

import { beforeEach, describe, expect, test } from "vitest";

process.env.KEEL_DB_CONN_TYPE = "pg";
process.env.KEEL_DB_CONN = `postgresql://postgres:postgres@localhost:5432/functions-runtime`;

let permissions;
let ctx = {};
let db = getDatabase();

describe("explicit", () => {
  beforeEach(() => {
    permissions = new Permissions();
  });

  test("explicitly allowing execution", () => {
    expect(permissions.getState()).toEqual(PERMISSION_STATE.UNKNOWN);

    permissions.allow();

    expect(permissions.getState()).toEqual(PERMISSION_STATE.PERMITTED);
  });

  test("explicitly denying execution", () => {
    expect(permissions.getState()).toEqual(PERMISSION_STATE.UNKNOWN);

    expect(() => permissions.deny()).toThrowError(PermissionError);

    expect(permissions.getState()).toEqual(PERMISSION_STATE.UNPERMITTED);
  });
});

describe("check", () => {
  const functionName = "createPerson";

  test("check - success", async () => {
    const permissionRule1 = (records, ctx, db) => {
      // Only allow names starting with Adam
      return records.every((r) => r.name.startsWith("Adam"));
    };

    const rows = [
      {
        id: "123",
        name: "Adam Bull",
      },
      {
        id: "234",
        name: "Adam Lambert",
      },
    ];

    await expect(
      checkBuiltInPermissions({
        rows,
        ctx,
        db,
        functionName,
        permissions: [permissionRule1],
      })
    ).resolves.ok;
  });

  test("check - failure", async () => {
    // only allow Petes
    const permissionRule1 = (records, ctx, db) => {
      return records.every((r) => r.name === "Pete");
    };

    const rows = [
      {
        id: "123",
        name: "Adam", // this one will cause an error to be thrown because Adam is not Pete
      },
      {
        id: "234",
        name: "Pete",
      },
    ];

    await expect(
      checkBuiltInPermissions({
        rows,
        ctx,
        db,
        functionName,
        permissions: [permissionRule1],
      })
    ).rejects.toThrow();
  });

  test("with a single row", async () => {
    const permissionRule1 = (records, ctx, db) => {
      // Only allow names starting with Adam
      return records.every((r) => r.name.startsWith("Adam"));
    };

    const rows = {
      id: "123",
      name: "Adam Bull",
    };

    await expect(
      checkBuiltInPermissions({
        rows,
        ctx,
        db,
        functionName,
        permissions: [permissionRule1],
      })
    ).resolves.ok;
  });
});
