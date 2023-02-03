const { ModelAPI } = require("./ModelAPI");
const { RequestHeaders } = require("./RequestHeaders");
const { handleRequest } = require("./handleRequest");
const KSUID = require("ksuid");
const { getDatabase } = require("./database");

module.exports = {
  ModelAPI,
  RequestHeaders,
  handleRequest,
  getDatabase,
  ksuid() {
    return KSUID.randomSync().string;
  },
};
