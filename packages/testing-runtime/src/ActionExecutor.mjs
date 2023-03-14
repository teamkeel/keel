import jwt from "jsonwebtoken";

function getHeaders(identity, authToken) {
  const headers = { "Content-Type": "application/json" };

  // An Identity instance is provided make a JWT
  if (identity !== null) {
    headers["Authorization"] =
      "Bearer " +
      jwt.sign(
        {
          id: identity.id,
        },
        // TODO: make this an env var
        "test",
        { algorithm: "HS256", expiresIn: 60 * 60 * 24 }
      );
  }

  // If an auth token is provided that can be sent as-is
  if (authToken !== null) {
    headers["Authorization"] = "Bearer " + authToken;
  }

  return headers;
}

export class ActionExecutor {
  constructor(props) {
    this._identity = props.identity || null;
    this._authToken = props.authToken || null;

    // Return a proxy which will return a bound version of the
    // _execute method for any unknown properties. This creates
    // the actions API we want but in a dynamic way without needing
    // codegen. We then generate the right type definitions for
    // this class in the @teamkeel/testing package.
    return new Proxy(this, {
      get(target, prop) {
        const v = Reflect.get(...arguments);
        if (v !== undefined) {
          return v;
        }
        return target._execute.bind(target, prop);
      },
    });
  }
  withIdentity(i) {
    return new ActionExecutor({ identity: i });
  }
  withAuthToken(t) {
    return new ActionExecutor({ authToken: t });
  }
  async graphql(params) {
    const headers = getHeaders(this._identity, this._authToken);

    const res = await fetch(
      process.env.KEEL_TESTING_ACTIONS_API_URL + "/graphql",
      {
        method: "POST",
        body: JSON.stringify(params),
        headers,
      }
    );
    const body = await res.json();
    if (body.errors) {
      return {
        errors: body.errors,
      };
    }
    return {
      data: body.data,
    };
  }
  _execute(method, params) {
    const headers = getHeaders(this._identity, this._authToken);

    // Use the HTTP JSON API as that returns more friendly errors than
    // the JSON-RPC API.
    return fetch(process.env.KEEL_TESTING_ACTIONS_API_URL + "/json/" + method, {
      method: "POST",
      body: JSON.stringify(params),
      headers,
    }).then((r) => {
      if (r.status !== 200) {
        // For non-200 first read the response as text
        return r.text().then((t) => {
          let d;
          try {
            d = JSON.parse(t);
          } catch (e) {
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
      return r.json();
    });
  }
}
