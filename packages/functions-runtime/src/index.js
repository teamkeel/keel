const { ModelAPI } = require("./ModelAPI");
const { RequestHeaders } = require("./RequestHeaders");
const { handleRequest } = require("./handleRequest");
const KSUID = require("ksuid");
const { useDatabase } = require("./database");
const {
  Permissions,
  PERMISSION_STATE,
  checkBuiltInPermissions,
} = require("./permissions");
const tracing = require("./tracing");

module.exports = {
  ModelAPI,
  RequestHeaders,
  handleRequest,
  useDatabase,
  Permissions,
  PERMISSION_STATE,
  checkBuiltInPermissions,
  tracing,
  ksuid() {
    return KSUID.randomSync().string;
  },
};