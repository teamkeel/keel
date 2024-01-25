export class Core {
  constructor(
    private config: RequestConfig,
    private refreshToken: TokenStore = new LocalStateStore()
  ) {
    this.auth.refreshSession();
  }

  #session: AccessTokenSession | null = null;

  ctx = {
    /**
     * @deprecated This has been deprecated in favour of using the APIClient.auth which handles sessions implicitly
     */
    token: "",
    /**
     * @deprecated This has been deprecated in favour of APIClient.auth.isAuthenticated()
     */
    isAuthenticated: false,
  };

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
    /**
     * @deprecated This has been deprecated in favour of the APIClient.auth authenticate helper functions
     */
    setToken: (value: string): Core => {
      this.ctx.token = value;
      this.ctx.isAuthenticated = true;
      return this;
    },
    /**
     * @deprecated This has been deprecated in favour of APIClient.auth.logout()
     */
    clearToken: (): Core => {
      this.ctx.token = "";
      this.ctx.isAuthenticated = false;
      return this;
    },
    rawRequest: async <T>(action: string, body: any): Promise<APIResult<T>> => {
      // If necessary, refresh the expired session before calling the action
      await this.auth.isAuthenticated();

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
              ...(this.#session || this.ctx.token != ""
                ? {
                    Authorization:
                      "Bearer " + this.#session?.token ?? this.ctx.token,
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
     * Returns the list of supported authentication providers and their SSO login URLs.
     */
    providers: async (): Promise<Provider[]> => {
      let url = new URL(this.config.baseUrl);
      const result = await globalThis.fetch(url.origin + "/auth/providers", {
        method: "GET",
        cache: "no-cache",
        headers: {
          "content-type": "application/json",
        },
      });

      if (result.ok) {
        const rawJson = await result.text();
        return JSON.parse(rawJson);
      } else {
        throw new Error(
          "unexpected status code response from /auth/providers: " +
            result.status
        );
      }
    },

    /**
     * Returns true if the session has not expired. If expired, it will attempt to refresh the session from the authentication server.
     */
    isAuthenticated: async () => {
      // If there is no session, then we don't attempt to refresh since
      // the client was not authenticated in the first place.
      if (!this.#session) {
        return false;
      }

      // Consider a token expired EXPIRY_BUFFER_IN_MS earlier than its real expiry time
      const isExpired =
        Date.now() > this.#session!.expiresAt.getTime() - EXPIRY_BUFFER_IN_MS;

      if (isExpired) {
        return await this.auth.refreshSession();
      }

      return true;
    },

    /**
     * Authenticate with an ID token.
     */
    authenticateWithIdToken: async (idToken: string) => {
      const req: TokenExchangeGrant = {
        grant: "token_exchange",
        subjectToken: idToken,
      };

      await this.auth.requestToken(req);
    },

    /**
     * Authenticate with Single Sign On using the auth code received from a successful SSO flow.
     */
    authenticateWithSingleSignOn: async (code: string) => {
      const req: AuthorizationCodeGrant = {
        grant: "authorization_code",
        code: code,
      };

      await this.auth.requestToken(req);
    },

    /**
     * Forcefully refreshes the session with the authentication server and returns true if authenticated.
     */
    refresh: async () => {
      const refreshToken = this.refreshToken.get();

      if (!refreshToken) {
        return false;
      }

      const authorised = await this.auth.requestToken({
        grant: "refresh_token",
        refreshToken: refreshToken,
      });

      if (!authorised) {
        this.refreshToken.set(null);
      }

      return authorised;
    },

    /**
     * Logs out the session on the client and also attempts to revoke the refresh token with the authentication server.
     */
    logout: async () => {
      const refreshToken = this.refreshToken.get();

      this.#session = null;
      this.refreshToken.set(null);

      if (refreshToken) {
        let url = new URL(this.config.baseUrl);
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
    requestToken: async (req: TokenGrant) => {
      let body = null;
      switch (req.grant) {
        case "token_exchange":
          body = {
            subject_token: req.subjectToken,
          };
          break;
        case "authorization_code":
          body = {
            code: req.code,
          };
          break;
        case "refresh_token":
          body = {
            refresh_token: req.refreshToken,
          };
          break;
        default:
          throw new Error(
            "Unknown grant type. We currently support 'authorization_code', 'token_exchange', and 'refresh_token' grant types. Please use one of those. For more info, please refer to the docs at https://docs.keel.so/authentication/endpoints#parameters"
          );
      }

      let url = new URL(this.config.baseUrl);
      const result = await globalThis.fetch(url.origin + "/auth/token", {
        method: "POST",
        cache: "no-cache",
        headers: {
          accept: "application/json",
          "content-type": "application/json",
        },
        body: JSON.stringify({
          grant_type: req.grant,
          ...body,
        }),
      });

      if (result.ok) {
        const rawJson = await result.text();
        const data = JSON.parse(rawJson);

        const expiresAt = new Date(Date.now() + data.expires_in * 1000);
        this.refreshToken.set(data.refresh_token);
        this.#session = { token: data.access_token, expiresAt: expiresAt };

        return true;
      } else if (result.status == 401) {
        return false;
      } else {
        const resp = await result.json();
        throw new TokenError(resp.error, resp.error_description);
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

class LocalStateStore {
  private token: string | null = null;
  get = () => {
    return this.token;
  };
  set = (token: string) => {
    this.token = token;
  };
}