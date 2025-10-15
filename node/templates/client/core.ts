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
      try {
        // If necessary, refresh the expired session before calling the action
        const isAuth = await this.auth.isAuthenticated();
        if (isAuth.error) {
          return { error: isAuth.error };
        }

        const token = this.auth.accessToken.get();

        let tz;
        try {
          tz = Intl.DateTimeFormat().resolvedOptions().timeZone;
        } catch {}

        const result = await globalThis.fetch(
          stripTrailingSlash(this.config.baseUrl) + "/json/" + action,
          {
            method: "POST",
            cache: "no-cache",
            headers: {
              accept: "application/json",
              "content-type": "application/json",
              ...(tz ? { "time-zone": tz } : {}),
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
          // Use text() to get the response body - this will automatically
          // decompress gzip responses when Content-Encoding: gzip is set
          try {
            const rawJson = await result.text();
            const data = JSON.parse(rawJson, reviver);

            return {
              data,
            };
          } catch (err) {
            console.error("[CLIENT ERROR] Failed to parse response:", err);
            console.error("[CLIENT ERROR] Error details:", err.message, err.cause);
            throw err;
          }
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
    accessToken: this.config?.accessTokenStore || new InMemoryStore(),

    /**
     * Get or set the refresh token from the configured token store.
     */
    refreshToken: this.config?.refreshTokenStore || new InMemoryStore(),

    /**
     * A promise that resolves when the session is refreshed.
     */
    refreshingPromise: undefined as
      | Promise<APIResult<AuthenticationResponse>>
      | undefined,

    /**
     * Returns data field set to the list of supported authentication providers and their SSO login URLs.
     * Returns error field if an unexpected error occurred.
     */
    providers: async (): Promise<APIResult<Provider[]>> => {
      const url = new URL(this.config.baseUrl);
      const result = await globalThis.fetch(url.origin + "/auth/providers", {
        method: "GET",
        cache: "no-cache",
        headers: {
          "content-type": "application/json",
        },
      });

      if (result.ok) {
        return { data: await result.json() };
      } else {
        return {
          error: {
            message:
              "unexpected status code response from /auth/providers: " +
              result.status,
            type: "unknown",
          },
        };
      }
    },

    /**
     * Returns data field set to the time at which the session will expire.
     * Returns error field if an unexpected error occurred.
     */
    expiresAt: (): APIResult<Date | null> => {
      const token = this.auth.accessToken.get();

      if (!token) {
        return { data: null };
      }

      let payload;
      try {
        const base64Payload = token.split(".")[1];
        payload = atob(base64Payload);
      } catch (e) {
        return {
          error: {
            message: "jwt token could not be parsed",
            type: "unknown",
          },
        };
      }

      var obj = JSON.parse(payload);
      if (obj !== null && typeof obj === "object") {
        const { exp } = obj as {
          exp: number;
        };

        return { data: new Date(exp * 1000) };
      }

      return {
        error: {
          message: "jwt token could not be parsed from json",
          type: "unknown",
        },
      };
    },

    /**
     * Returns data field set to true if the session has not expired. If expired, it will attempt to refresh the session from the authentication server.
     * Returns error field if an unexpected error occurred.
     */
    isAuthenticated: async (): Promise<APIResult<boolean>> => {
      // If there is no access token, then attempt to refresh it.

      if (!this.auth.accessToken.get()) {
        const res = await this.auth.refresh();
        if (res.error) {
          return {
            error: res.error,
          };
        }

        return res;
      }

      // Consider a token expired EXPIRY_BUFFER_IN_MS earlier than its real expiry time
      const expiresAt = this.auth.expiresAt();
      if (expiresAt.error) {
        return {
          error: expiresAt.error,
        };
      }

      const isExpired =
        expiresAt.data != null
          ? Date.now() > expiresAt.data.getTime() - EXPIRY_BUFFER_IN_MS
          : false;

      if (isExpired) {
        return await this.auth.refresh();
      }

      return {
        data: true,
      };
    },

    /**
     * Authenticates with email and password flow and, if successful, returns data field with result of the authentication.
     * Returns error field if an error occurred.
     */
    authenticateWithPassword: async (
      input: PasswordFlowInput
    ): Promise<APIResult<AuthenticationResponse>> => {
      const req: PasswordGrant = {
        grant_type: "password",
        username: input.email,
        password: input.password,
        create_if_not_exists: input.createIfNotExists,
      };

      return await this.auth.requestToken(req);
    },

    /**
     * Authenticates with the ID Token flow and, if successful, returns data field with result of the authentication.
     * Returns error field if an error occurred.
     */
    authenticateWithIdToken: async (
      input: IDTokenFlowInput
    ): Promise<APIResult<AuthenticationResponse>> => {
      const req: TokenExchangeGrant = {
        grant_type: "token_exchange",
        subject_token: input.idToken,
        create_if_not_exists: input.createIfNotExists,
      };

      return await this.auth.requestToken(req);
    },

    /**
     * Authenticates with the Single Sign-On flow and, if successful, returns data field with result of the authentication.
     * Returns error field if an error occurred.
     */
    authenticateWithSingleSignOn: async (
      input: SingleSignOnFlowInput
    ): Promise<APIResult<AuthenticationResponse>> => {
      const req: AuthorizationCodeGrant = {
        grant_type: "authorization_code",
        code: input.code,
      };

      return await this.auth.requestToken(req);
    },

    /**
     * Forcefully refreshes the session with the authentication server, and returns data field set to true if the identity is still authenticated.
     * Return true if successfully authenticated.
     */
    refresh: async (): Promise<APIResult<boolean>> => {
      const refreshToken = this.auth.refreshToken.get();

      if (!refreshToken) {
        return {
          data: false,
        };
      }

      // If refreshing already, wait for the existing refreshing promisee
      if (!this.auth.refreshingPromise) {
        this.auth.refreshingPromise = this.auth.requestToken({
          grant_type: "refresh_token",
          refresh_token: refreshToken,
        });
      }

      const authResponse = await this.auth.refreshingPromise;

      this.auth.refreshingPromise = undefined;

      if (authResponse.error) {
        return { error: authResponse.error };
      }

      return {
        data: true,
      };
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
    requestToken: async (
      req: TokenRequest
    ): Promise<APIResult<AuthenticationResponse>> => {
      try {
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

          return {
            data: {
              identityCreated: data.identity_created,
            },
          };
        } else {
          this.auth.accessToken.set(null);
          this.auth.refreshToken.set(null);

          const requestId = result.headers.get("X-Amzn-Requestid") || undefined;

          let errorMessage = "unknown error";

          try {
            const resp = await result.json();
            errorMessage = resp.error_description;
          } catch (error) {}

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
