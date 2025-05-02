interface RequestHeadersMap {
  [key: string]: string;
}

export class RequestHeaders {
  private _headers: Headers;

  /**
   * @param {{Object.<string, string>}} requestHeaders Map of request headers submitted from the client
   */

  constructor(requestHeaders: RequestHeadersMap) {
    this._headers = new Headers(requestHeaders);
  }

  get: Headers["get"] = (key: string) => {
    return this._headers.get(key);
  };

  has: Headers["has"] = (key: string) => {
    return this._headers.has(key);
  };
}
