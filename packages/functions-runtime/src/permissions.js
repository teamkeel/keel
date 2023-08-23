const { AsyncLocalStorage } = require("async_hooks");
const { PermissionError } = require("./errors");

const PERMISSION_STATE = {
  UNKNOWN: "unknown",
  PERMITTED: "permitted",
  UNPERMITTED: "unpermitted",
};

// withPermissions sets the initial permission state from the go runtime in the AsyncLocalStorage so consumers further down the hierarchy can read or mutate the state
// at will
const withPermissions = async (initialValue, cb) => {
  const permissions = new Permissions();

  return await permissionsApiInstance.run({ permitted: initialValue }, () => {
    return cb({ getPermissionState: permissions.getState });
  });
};

const permissionsApiInstance = new AsyncLocalStorage();

class Permissions {
  // The Go runtime performs role based permission rule checks prior to calling the functions
  // runtime, so the status could already be granted. If already granted, then we need to inherit that permission state as the state is later used to decide whether to run in process permission checks
  // TLDR if a role based permission is relevant and it is granted, then it is effectively the same as the end user calling api.permissions.allow() explicitly in terms of behaviour.

  allow() {
    const store = (permissionsApiInstance.getStore().permitted = true);
  }

  deny() {
    // if a user is explicitly calling deny() then we want to throw an error
    // so that any further execution of the custom function stops abruptly
    permissionsApiInstance.getStore().permitted = false;

    throw new PermissionError();
  }

  getState() {
    const permitted = permissionsApiInstance.getStore().permitted;

    switch (true) {
      case permitted === false:
        return PERMISSION_STATE.UNPERMITTED;
      case permitted === null:
        return PERMISSION_STATE.UNKNOWN;
      case permitted === true:
        return PERMISSION_STATE.PERMITTED;
    }
  }
}

const checkBuiltInPermissions = async ({
  rows,
  permissionFns,
  ctx,
  db,
  functionName,
}) => {
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

module.exports.checkBuiltInPermissions = checkBuiltInPermissions;
module.exports.PermissionError = PermissionError;
module.exports.PERMISSION_STATE = PERMISSION_STATE;
module.exports.Permissions = Permissions;
module.exports.withPermissions = withPermissions;
module.exports.permissionsApiInstance = permissionsApiInstance;
