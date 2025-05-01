/** @type {import('node:http').Headers} */
const Headers = require("node:http").Headers;

class RequestHeaders {
  /**
   * @param {{Object.<string, string>}} requestHeaders Map of request headers submitted from the client
   */

  constructor(requestHeaders) {
    this._headers = new Headers(requestHeaders);
  }

  get(key) {
    return this._headers.get(key);
  }

  has(key) {
    return this._headers.has(key);
  }
}

module.exports = {
  RequestHeaders,
};
