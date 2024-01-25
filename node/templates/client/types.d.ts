type RequestHeaders = globalThis.Record<string, string>;

// Refresh the token 60 seconds before it expires
const EXPIRY_BUFFER_IN_MS = 1000;

export type RequestConfig = {
  baseUrl: string;
  headers?: RequestHeaders;
};

export interface TokenStore {
  set(token: string | null): void;
  get(): string | null;
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

// Result types

export type APIResult<T> = Result<T, APIError>;

type Data<T> = {
  data: T;
  error?: never;
};

type Err<U> = {
  data?: never;
  error: U;
};

type Result<T, U> = NonNullable<Data<T> | Err<U>>;

// Error types

/* 400 */
type BadRequestError = {
  type: "bad_request";
  message: string;
  requestId?: string;
};

/* 401 */
type UnauthorizedError = {
  type: "unauthorized";
  message: string;
  requestId?: string;
};

/* 403 */
type ForbiddenError = {
  type: "forbidden";
  message: string;
  requestId?: string;
};

/* 404 */
type NotFoundError = {
  type: "not_found";
  message: string;
  requestId?: string;
};

/* 500 */
type InternalServerError = {
  type: "internal_server_error";
  message: string;
  requestId?: string;
};

/* Unhandled/unexpected errors */
type UnknownError = {
  type: "unknown";
  message: string;
  error?: unknown;
  requestId?: string;
};

export type APIError =
  | UnauthorizedError
  | ForbiddenError
  | NotFoundError
  | BadRequestError
  | InternalServerError
  | UnknownError;

// Auth

export type Provider = {
  name: string;
  type: string;
  authorizeUrl: string;
};

export type AccessTokenSession = {
  token: string;
  expiresAt: Date;
};

export type TokenExchangeGrant = {
  grant: "token_exchange";
  subjectToken: string;
};

export type AuthorizationCodeGrant = {
  grant: "authorization_code";
  code: string;
};

export type RefreshGrant = {
  grant: "refresh_token";
  refreshToken: string;
};

export type TokenGrant =
  | TokenExchangeGrant
  | AuthorizationCodeGrant
  | RefreshGrant;

export class TokenError extends Error {
  errorDescription: string;
  constructor(error: string, errorDescription: string) {
    super();
    this.message = error;
    this.errorDescription = errorDescription;
  }
}

export type AuthenticateParams = SsoLogin | IdToken | UsernamePassword;

export type SsoLogin = {
  kind: "sso_login";
  code: string;
};

export type IdToken = {
  kind: "id_token";
  idToken: string;
};

export type UsernamePassword = {
  kind: "username_password";
  username: string;
  password: string;
};
