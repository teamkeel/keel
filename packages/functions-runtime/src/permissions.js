class PermitError extends Error {}

const PERMISSION_STATE = {
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

  async check(rows) {
    throw new Error("Not implemented");
  }

  allow() {
    this.state.permitted = true;
  }

  deny() {
    // if a user is explicitly calling deny() then we want to throw an error
    // so that any further execution of the custom function stops abruptly
    // we don't need to explicitly set pending to false as the error will have been thrown
    // so we know an action has been taken
    throw new PermitError();
  }

  getState() {
    switch (true) {
      // this will cover both permitted being false, and null (initial state)
      case !this.state.permitted:
        return PERMISSION_STATE.UNPERMITTED;
      default:
        return PERMISSION_STATE.PERMITTED;
    }
  }
}

module.exports.PermitError = PermitError;
module.exports.PERMISSION_STATE = PERMISSION_STATE;
module.exports.Permissions = Permissions;
