class PermissionError extends Error {}

const PERMISSION_STATE = {
  UNKNOWN: "unknown",
  PERMITTED: "permitted",
  UNPERMITTED: "unpermitted",
};

class Permissions {
  constructor() {
    this.state = {
      // permitted starts off as null to indicate that the end user
      // hasn't explicitly marked a function execution as permitted yet
      permitted: null,
    };
  }

  allow() {
    this.state.permitted = true;
  }

  deny() {
    // if a user is explicitly calling deny() then we want to throw an error
    // so that any further execution of the custom function stops abruptly
    this.state.permitted = false;

    throw new PermissionError();
  }

  getState() {
    switch (true) {
      case this.state.permitted === false:
        return PERMISSION_STATE.UNPERMITTED;
      case this.state.permitted === null:
        return PERMISSION_STATE.UNKNOWN;
      case this.state.permitted === true:
        return PERMISSION_STATE.PERMITTED;
    }
  }
}

const checkBuiltInPermissions = async ({
  rows,
  permissions,
  ctx,
  db,
  functionName,
}) => {
  // rows can actually just be a single record too so we need to wrap it
  if (!Array.isArray(rows)) {
    rows = [rows];
  }

  for (const permissionFn of permissions) {
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
