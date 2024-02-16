type RequestHeaders = globalThis.Record<string, string>;

// Refresh the token EXPIRY_BUFFER_IN_MS seconds before it expires
const EXPIRY_BUFFER_IN_MS = 60000;

export type Config = {
  baseUrl: string;
  headers?: RequestHeaders;
  refreshToken?: TokenStore;
  accessToken?: TokenStore;
};

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

export interface TokenStore {
  set(token: string | null): void;
  get(): string | null;
}

export type Provider = {
  name: string;
  type: string;
  authorizeUrl: string;
};

export type PasswordGrant = {
  grant_type: "password";
  username: string;
  password: string;
};

export type TokenExchangeGrant = {
  grant_type: "token_exchange";
  subject_token: string;
};

export type AuthorizationCodeGrant = {
  grant_type: "authorization_code";
  code: string;
};

export type RefreshGrant = {
  grant_type: "refresh_token";
  refresh_token: string;
};

export type TokenRequest =
  | PasswordGrant
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
