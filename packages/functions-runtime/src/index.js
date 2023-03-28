const { ModelAPI } = require("./ModelAPI");
const { RequestHeaders } = require("./RequestHeaders");
const { handleRequest } = require("./handleRequest");
const KSUID = require("ksuid");
const { getDatabase } = require("./database");
const {
  Permissions,
  PERMISSION_STATE,
  checkBuiltInPermissions,
} = require("./permissions");

module.exports = {
  ModelAPI,
  RequestHeaders,
  handleRequest,
  getDatabase,
  Permissions,
  PERMISSION_STATE,
  checkBuiltInPermissions,
  ksuid() {
    return KSUID.randomSync().string;
  },
};
