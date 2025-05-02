import { AsyncLocalStorage } from "async_hooks";
import { PermissionError } from "./errors";

export const PERMISSION_STATE = {
  UNKNOWN: "unknown",
  PERMITTED: "permitted",
  UNPERMITTED: "unpermitted",
} as const;

export type PermissionState =
  (typeof PERMISSION_STATE)[keyof typeof PERMISSION_STATE];

/**
 * Permissions class for managing access control in the runtime
 */
export class Permissions {
  // The Go runtime performs role based permission rule checks prior to calling the functions
  // runtime, so the status could already be granted. If already granted, then we need to inherit that permission state as the state is later used to decide whether to run in process permission checks
  // TLDR if a role based permission is relevant and it is granted, then it is effectively the same as the end user calling api.permissions.allow() explicitly in terms of behaviour.

  /**
   * Explicitly permit access to an action
   */
  allow(): void {
    permissionsApiInstance.getStore()!.permitted = true;
  }

  /**
   * Explicitly deny access to an action
   */
  deny(): never {
    // if a user is explicitly calling deny() then we want to throw an error
    // so that any further execution of the custom function stops abruptly
    permissionsApiInstance.getStore()!.permitted = false;
    throw new PermissionError();
  }

  getState(): PermissionState {
    const permitted = permissionsApiInstance.getStore()!.permitted;

    switch (true) {
      case permitted === false:
        return PERMISSION_STATE.UNPERMITTED;
      case permitted === null:
        return PERMISSION_STATE.UNKNOWN;
      case permitted === true:
        return PERMISSION_STATE.PERMITTED;
      default:
        return PERMISSION_STATE.UNKNOWN;
    }
  }
}

interface PermissionStore {
  permitted: boolean | null;
}

const permissionsApiInstance = new AsyncLocalStorage<PermissionStore>();

interface PermissionCallback {
  getPermissionState: () => PermissionState;
}

// withPermissions sets the initial permission state from the go runtime in the AsyncLocalStorage so consumers further down the hierarchy can read or mutate the state
// at will
export const withPermissions = async <T>(
  initialValue: boolean | null,
  cb: (permissions: PermissionCallback) => Promise<T>
): Promise<T> => {
  const permissions = new Permissions();

  return await permissionsApiInstance.run({ permitted: initialValue }, () => {
    return cb({ getPermissionState: permissions.getState });
  });
};

interface CheckBuiltInPermissionsParams {
  rows: any[];
  permissionFns: Array<(rows: any[], ctx: any, db: any) => Promise<boolean>>;
  ctx: any;
  db: any;
  functionName: string;
}

export const checkBuiltInPermissions = async ({
  rows,
  permissionFns,
  ctx,
  db,
  functionName,
}: CheckBuiltInPermissionsParams): Promise<void> => {
  for (const permissionFn of permissionFns) {
    const result = await permissionFn(rows, ctx, db);
    // if any of the permission functions return true,
    // then we return early
    if (result) {
      return;
    }
  }

  throw new PermissionError(`Not permitted to access ${functionName}`);
};

export { PermissionError };
export { permissionsApiInstance };
