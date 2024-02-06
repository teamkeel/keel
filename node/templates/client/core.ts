export class Core {
  constructor(
    private config: RequestConfig,
    private getTokens = InMemoryTokenStore.getInstance().getTokens,
    private setTokens = InMemoryTokenStore.getInstance().setTokens
  ) {}

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

      const token = this.getTokens().accessToken ?? this.ctx.token;

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
              ...(token != "" && action != "authenticate"
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
      const token = this.getTokens().accessToken;

      if (!token) {
        return null;
      }

      const payload = Buffer.from(token.split(".")[1], "base64").toString(
        "utf8"
      );

      var obj = JSON.parse(payload);
      if (obj !== null && typeof obj === "object") {
        const { exp } = obj as {
          exp: number;
        };

        return new Date(exp * 1000);
      }

      throw new Error("jwt token could not be parsed");
    },

    /**
     * Returns true if the session has not expired. If expired, it will attempt to refresh the session from the authentication server.
     */
    isAuthenticated: async () => {
      // If there is no session, then we don't attempt to refresh since
      // the client was not authenticated in the first place.
      if (!this.getTokens().accessToken) {
        return false;
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
      const refreshToken = this.getTokens().refreshToken;

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
      const refreshToken = this.getTokens().refreshToken;

      this.setTokens(null, null);

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
        this.setTokens(data.access_token, data.refresh_token);

        return true;
      } else {
        this.setTokens(null, null);

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

class InMemoryTokenStore {
  private static instance: InMemoryTokenStore;
  private accessToken: string | null = null;
  private refreshToken: string | null = null;

  private constructor() {}

  public static getInstance(): InMemoryTokenStore {
    if (!InMemoryTokenStore.instance) {
      InMemoryTokenStore.instance = new InMemoryTokenStore();
    }
    return InMemoryTokenStore.instance;
  }

  getTokens = () => {
    return {
      accessToken: this.accessToken,
      refreshToken: this.refreshToken,
    };
  };

  setTokens = (accessToken: string | null, refreshToken: string | null) => {
    this.accessToken = accessToken;
    this.refreshToken = refreshToken;
  };
}
