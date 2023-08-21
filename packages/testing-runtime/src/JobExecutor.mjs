import jwt from "jsonwebtoken";

export class JobExecutor {
  constructor(props) {
    this._identity = props.identity || null;
    this._authToken = props.authToken || null;

    // Return a proxy which will return a bound version of the
    // _execute method for any unknown properties. This creates
    // the jobs API we want but in a dynamic way without needing
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
    return new JobExecutor({ identity: i });
  }
  withAuthToken(t) {
    return new JobExecutor({ authToken: t });
  }
  _execute(method, params) {
    const headers = { "Content-Type": "application/json" };

    // An Identity instance is provided make a JWT using the default private key
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
          issuer: "keel",
        });
    }

    // If an auth token is provided that can be sent as-is
    if (this._authToken !== null) {
      headers["Authorization"] = "Bearer " + this._authToken;
    }

    if (params?.scheduled) {
      headers["X-Trigger-Type"] = "scheduled";
    } else {
      headers["X-Trigger-Type"] = "manual";
    }

    return fetch(process.env.KEEL_TESTING_JOBS_URL + "/" + method, {
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

      return true;
    });
  }
}
