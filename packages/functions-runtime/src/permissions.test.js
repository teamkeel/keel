const {
  permissionsApiInstance,
  Permissions,
  PERMISSION_STATE,
  checkBuiltInPermissions,
} = require("./permissions");
import { useDatabase } from "./database";
import { beforeEach, describe, expect, test } from "vitest";
const { PermissionError } = require("./errors");

let permissions;
let ctx = {};
let db = useDatabase();

describe("explicit", () => {
  beforeEach(() => {
    permissions = new Permissions();
  });

  test("explicitly allowing execution", () => {
    wrapWithAsyncLocalStorage({ permitted: null }, () => {
      expect(permissions.getState()).toEqual(PERMISSION_STATE.UNKNOWN);

      permissions.allow();

      expect(permissions.getState()).toEqual(PERMISSION_STATE.PERMITTED);
    });
  });

  test("explicitly denying execution", () => {
    wrapWithAsyncLocalStorage({ permitted: null }, () => {
      expect(permissions.getState()).toEqual(PERMISSION_STATE.UNKNOWN);

      expect(() => permissions.deny()).toThrowError(PermissionError);

      expect(permissions.getState()).toEqual(PERMISSION_STATE.UNPERMITTED);
    });
  });
});

describe("prior state", () => {
  test("when the prior state is granted", () => {
    wrapWithAsyncLocalStorage(
      {
        permitted: true,
      },
      () => {
        expect(permissions.getState()).toEqual(PERMISSION_STATE.PERMITTED);
      }
    );
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
        permissionFns: [permissionRule1],
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
        permissionFns: [permissionRule1],
      })
    ).rejects.toThrow();
  });
});

function wrapWithAsyncLocalStorage(initialState, testFn) {
  permissionsApiInstance.run(initialState, () => {
    testFn();
  });
}
