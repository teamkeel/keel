const { ModelAPI } = require("./ModelAPI");
const { handleRequest } = require("./handleRequest");
const KSUID = require("ksuid");
const { getDatabase } = require("./database");

module.exports = {
  ModelAPI,
  handleRequest,
  getDatabase,
  ksuid() {
    return KSUID.randomSync().string;
  },
};
