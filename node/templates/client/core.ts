export class Core {
  constructor(private config: Config) {}

  client = {
    setHeaders: (headers: RequestHeaders): Core => {
      this.config.headers = headers;
      return this;
    },
    setHeader: (key: string, value: string): Core => {
      const { headers } = this.config;
      if (headers) {
        headers[key] = value;
      } else {
        this.config.headers = { [key]: value };
      }
      return this;
    },
    setBaseUrl: (value: string): Core => {
      this.config.baseUrl = value;
      return this;
    },
    rawRequest: async <T>(action: string, body: any): Promise<APIResult<T>> => {
      // If necessary, refresh the expired session before calling the action
      await this.auth.isAuthenticated();

      const token = this.auth.accessToken.get();

      try {
        const result = await globalThis.fetch(
          stripTrailingSlash(this.config.baseUrl) + "/json/" + action,
          {
            method: "POST",
            cache: "no-cache",
            headers: {
              accept: "application/json",
              "content-type": "application/json",
              ...this.config.headers,
              ...(token != null
                ? {
                    Authorization: "Bearer " + token,
                  }
                : {}),
            },
            body: JSON.stringify(body),
          }
        );

        if (result.status >= 200 && result.status < 299) {
          const rawJson = await result.text();
          const data = JSON.parse(rawJson, reviver);

          return {
            data,
          };
        }

        let errorMessage = "unknown error";

        try {
          const errorData: {
            message: string;
          } = await result.json();
          errorMessage = errorData.message;
        } catch (error) {}

        const requestId = result.headers.get("X-Amzn-Requestid") || undefined;

        const errorCommon = {
          message: errorMessage,
          requestId,
        };

        switch (result.status) {
          case 400:
            return {
              error: {
                ...errorCommon,
                type: "bad_request",
              },
            };
          case 401:
            return {
              error: {
                ...errorCommon,
                type: "unauthorized",
              },
            };
          case 403:
            return {
              error: {
                ...errorCommon,
                type: "forbidden",
              },
            };
          case 404:
            return {
              error: {
                ...errorCommon,
                type: "not_found",
              },
            };
          case 500:
            return {
              error: {
                ...errorCommon,
                type: "internal_server_error",
              },
            };

          default:
            return {
              error: {
                ...errorCommon,
                type: "unknown",
              },
            };
        }
      } catch (error) {
        return {
          error: {
            type: "unknown",
            message: "unknown error",
            error,
          },
        };
      }
    },
  };

  auth = {
    /**
     * Get or set the access token from the configured token store.
     */
    accessToken: this.config.accessTokenStore || new InMemoryStore(),

    /**
     * Get or set the refresh token from the configured token store.
     */
    refreshToken: this.config.refreshTokenStore || new InMemoryStore(),

    /**
     * Returns the list of supported authentication providers and their SSO login URLs.
     */
    providers: async (): Promise<Provider[]> => {
      const url = new URL(this.config.baseUrl);
      const result = await globalThis.fetch(url.origin + "/auth/providers", {
        method: "GET",
        cache: "no-cache",
        headers: {
          "content-type": "application/json",
        },
      });

      if (result.ok) {
        return await result.json();
      } else {
        throw new Error(
          "unexpected status code response from /auth/providers: " +
            result.status
        );
      }
    },

    /**
     * Returns the time at which the session will expire.
     */
    expiresAt: (): Date | null => {
      const token = this.auth.accessToken.get();

      if (!token) {
        return null;
      }

      let payload;
      try {
        const base64Payload = token.split(".")[1];
        payload = atob(base64Payload);
      } catch (e) {
        throw new Error(
          "jwt token could not be parsed: " + (e as Error).message
        );
      }

      var obj = JSON.parse(payload);
      if (obj !== null && typeof obj === "object") {
        const { exp } = obj as {
          exp: number;
        };

        return new Date(exp * 1000);
      }

      throw new Error("jwt token could not be parsed from json");
    },

    /**
     * Returns true if the session has not expired. If expired, it will attempt to refresh the session from the authentication server.
     */
    isAuthenticated: async () => {
      // If there is no access token, then attempt to refresh it.
      if (!this.auth.accessToken.get()) {
        return await this.auth.refresh();
      }

      // Consider a token expired EXPIRY_BUFFER_IN_MS earlier than its real expiry time
      const expiresAt = this.auth.expiresAt();
      const isExpired =
        expiresAt != null
          ? Date.now() > expiresAt.getTime() - EXPIRY_BUFFER_IN_MS
          : false;

      if (isExpired) {
        return await this.auth.refresh();
      }

      return true;
    },

    /**
     * Authenticate with email and password.
     * Return true if successfully authenticated.
     */
    authenticateWithPassword: async (email: string, password: string) => {
      const req: PasswordGrant = {
        grant_type: "password",
        username: email,
        password: password,
      };

      await this.auth.requestToken(req);
    },

    /**
     * Authenticate with an ID token.
     * Return true if successfully authenticated.
     */
    authenticateWithIdToken: async (idToken: string) => {
      const req: TokenExchangeGrant = {
        grant_type: "token_exchange",
        subject_token: idToken,
      };

      await this.auth.requestToken(req);
    },

    /**
     * Authenticate with Single Sign On using the auth code received from a successful SSO flow.
     * Return true if successfully authenticated.
     */
    authenticateWithSingleSignOn: async (code: string) => {
      const req: AuthorizationCodeGrant = {
        grant_type: "authorization_code",
        code: code,
      };

      await this.auth.requestToken(req);
    },

    /**
     * Forcefully refreshes the session with the authentication server.
     * Return true if successfully authenticated.
     */
    refresh: async () => {
      const refreshToken = this.auth.refreshToken.get();

      if (!refreshToken) {
        return false;
      }

      const authorised = await this.auth.requestToken({
        grant_type: "refresh_token",
        refresh_token: refreshToken,
      });

      return authorised;
    },

    /**
     * Logs out the session on the client and also attempts to revoke the refresh token with the authentication server.
     */
    logout: async () => {
      const refreshToken = this.auth.refreshToken.get();

      this.auth.accessToken.set(null);
      this.auth.refreshToken.set(null);

      if (refreshToken) {
        const url = new URL(this.config.baseUrl);
        await globalThis.fetch(url.origin + "/auth/revoke", {
          method: "POST",
          cache: "no-cache",
          headers: {
            accept: "application/json",
            "content-type": "application/json",
          },
          body: JSON.stringify({
            token: refreshToken,
          }),
        });
      }
    },

    /**
     * Creates or refreshes a session with a token request at the authentication server.
     */
    requestToken: async (req: TokenRequest) => {
      const url = new URL(this.config.baseUrl);
      const result = await globalThis.fetch(url.origin + "/auth/token", {
        method: "POST",
        cache: "no-cache",
        headers: {
          accept: "application/json",
          "content-type": "application/json",
        },
        body: JSON.stringify(req),
      });

      if (result.ok) {
        const data = await result.json();

        this.auth.accessToken.set(data.access_token);
        this.auth.refreshToken.set(data.refresh_token);

        return true;
      } else {
        this.auth.accessToken.set(null);
        this.auth.refreshToken.set(null);

        if (result.status == 401) {
          return false;
        } else if (result.status == 400) {
          const resp = await result.json();
          throw new TokenError(resp.error, resp.error_description);
        } else {
          throw new Error(
            "unexpected status code " +
              result.status +
              " when requesting a token"
          );
        }
      }
    },
  };
}

const stripTrailingSlash = (str: string) => {
  if (!str) return str;
  return str.endsWith("/") ? str.slice(0, -1) : str;
};

const RFC3339 =
  /^(?:\d{4}-(?:0[1-9]|1[0-2])-(?:0[1-9]|[12][0-9]|3[01]))?(?:[T\s](?:[01]\d|2[0-3]):[0-5]\d(?::[0-5]\d)?(?:\.\d+)?(?:[Zz]|[+-](?:[01]\d|2[0-3]):?[0-5]\d)?)?$/;
function reviver(key: any, value: any) {
  // Convert any ISO8601/RFC3339 strings to dates
  if (value && typeof value === "string" && RFC3339.test(value)) {
    return new Date(value);
  }
  return value;
}

export class InMemoryStore implements TokenStore {
  private token: string | null = null;

  public constructor() {}

  get = () => {
    return this.token;
  };

  set = (token: string | null): void => {
    this.token = token;
  };
}
