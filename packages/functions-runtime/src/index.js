const { RequestHeaders } = require("./RequestHeaders");
const { handleRequest } = require("./handleRequest");
const KSUID = require("ksuid");
const { getDatabase } = require("./database");
const {
  Permissions,
  PERMISSION_STATE,
  checkBuiltInPermissions,
} = require("./permissions");
const tracing = require("./tracing");

module.exports = {
  RequestHeaders,
  handleRequest,
  getDatabase,
  Permissions,
  PERMISSION_STATE,
  checkBuiltInPermissions,
  tracing,
  ksuid() {
    return KSUID.randomSync().string;
  },
};
