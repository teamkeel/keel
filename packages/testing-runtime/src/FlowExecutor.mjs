import { parseInputs, parseOutputs, reviver } from "./parsing.mjs";

export class FlowExecutor {
  constructor(props) {
    this._flowUrl = process.env.KEEL_TESTING_FLOWS_API_URL + "/" + props.name;
    this._name = props.name;
    this._identity = props.identity || null;
    this._authToken = props.authToken || null;
  }

  withIdentity(i) {
    return new FlowExecutor({
      name: this._name,
      identity: i,
      apiBaseUrl: this._apiBaseUrl,
      parseJsonResult: this._parseJsonResult,
    });
  }
  withAuthToken(t) {
    return new FlowExecutor({
      name: this._name,
      authToken: t,
      apiBaseUrl: this._apiBaseUrl,
      parseJsonResult: this._parseJsonResult,
    });
  }

  headers() {
    const headers = { "Content-Type": "application/json" };

    // An Identity instance is provided make a JWT
    if (this._identity !== null) {
      const base64pk = process.env.KEEL_DEFAULT_PK;
      let privateKey = undefined;

      if (base64pk) {
        privateKey = Buffer.from(base64pk, "base64").toString("utf8");
      }

      headers["Authorization"] =
        "Bearer " +
        jwt.sign({}, privateKey, {
          algorithm: privateKey ? "RS256" : "none",
          expiresIn: 60 * 60 * 24,
          subject: this._identity.id,
          issuer: "https://keel.so",
        });
    }

    // If an auth token is provided that can be sent as-is
    if (this._authToken !== null) {
      headers["Authorization"] = "Bearer " + this._authToken;
    }

    return headers;
  }

  async start(inputs) {
    return parseInputs(inputs).then((parsed) => {
      // Use the HTTP JSON API as that returns more friendly errors than
      // the JSON-RPC API.
      return fetch(this._flowUrl, {
        method: "POST",
        body: JSON.stringify(parsed),
        headers: this.headers(),
      }).then(handleResponse);
    });
  }

  async get(id) {
    return fetch(this._flowUrl + "/" + id, {
      method: "GET",
      headers: this.headers(),
    }).then(handleResponse);
  }

  async cancel(id) {
    return fetch(this._flowUrl + "/" + id + "/cancel", {
      method: "POST",
      headers: this.headers(),
    }).then(handleResponse);
  }

  async putStepValues(id, stepId, values, action) {
    let url = this._flowUrl + "/" + id + "/" + stepId;

    if (action) {
      // If an action is provided then we need to add it to the URL
      const queryString = new URLSearchParams({ action }).toString();
      url = `${url}?${queryString}`;
    }

    return await fetch(url, {
      method: "PUT",
      body: JSON.stringify(values),
      headers: this.headers(),
    }).then(handleResponse);
  }

  async callback(id, stepId, element, callbackName, values)  {
    let url = this._flowUrl + "/" + id + "/" + stepId + "/callback?element=" + element + "&callback=" + callbackName;
    
    return await fetch(url, {
      method: "POST",
      body: JSON.stringify(values),
      headers: this.headers(),
    }).then(handleResponse);
  }

  async untilFinished(id, timeout = 5000) {
    const startTime = Date.now();

    while (true) {
      if (Date.now() - startTime > timeout) {
        throw new Error(
          `timed out waiting for flow run to reach a completed state after ${timeout}ms`
        );
      }

      const flow = await this.get(id);

      if (flow.status === "COMPLETED" || flow.status === "FAILED") {
        return flow;
      }

      await new Promise((resolve) => setTimeout(resolve, 100));
    }
  }

  async untilAwaitingInput(id, timeout = 5000) {
    const startTime = Date.now();

    while (true) {
      if (Date.now() - startTime > timeout) {
        throw new Error(
          `timed out waiting for flow run to reach a completed state after ${timeout}ms`
        );
      }

      const flow = await this.get(id);

      if (flow.status === "AWAITING_INPUT") {
        return flow;
      }

      await new Promise((resolve) => setTimeout(resolve, 100));
    }
  }
}

function handleResponse(r) {
  if (r.status !== 200) {
    // For non-200 first read the response as text
    return r.text().then((t) => {
      let d;
      try {
        d = JSON.parse(t);
      } catch (e) {
        if ("DEBUG" in process.env) {
          console.log(e);
        }
        // If JSON parsing fails then throw an error with the
        // response text as the message
        throw new Error(t);
      }
      // Otherwise throw the parsed JSON error response
      // We override toString as otherwise you get expect errors like:
      //   `expected to resolve but rejected with "[object Object]"`
      Object.defineProperty(d, "toString", {
        value: () => t,
        enumerable: false,
      });
      throw d;
    });
  }

  return r.text().then((t) => {
    const response = JSON.parse(t, reviver);
    response.input = parseOutputs(response.input);

    return response;
  });
}
