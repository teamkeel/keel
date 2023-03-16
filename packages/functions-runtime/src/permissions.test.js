import { Permissions, PERMISSION_STATE, PermissionError } from "./permissions";

import { beforeEach, expect, test } from "vitest";

let permissions;

beforeEach(() => {
  permissions = new Permissions();
});

test("explicitly allowing execution", () => {
  expect(permissions.getState()).toEqual(PERMISSION_STATE.UNPERMITTED);

  permissions.allow();

  expect(permissions.getState()).toEqual(PERMISSION_STATE.PERMITTED);
});

test("explicitly denying execution", () => {
  expect(permissions.getState()).toEqual(PERMISSION_STATE.UNPERMITTED);

  expect(() => permissions.deny()).toThrowError(PermissionError);

  expect(permissions.getState()).toEqual(PERMISSION_STATE.UNPERMITTED);
});

test("check", async () => {
  await expect(() => permissions.check()).rejects.toThrowError(
    "Not implemented"
  );
});
