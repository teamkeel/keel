type RequestHeaders = globalThis.Record<string, string>;

// Refresh the token EXPIRY_BUFFER_IN_MS seconds before it expires
const EXPIRY_BUFFER_IN_MS = 60000;

export type Config = {
  baseUrl: string;
  headers?: RequestHeaders;
  refreshTokenStore?: TokenStore;
  accessTokenStore?: TokenStore;
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

export type AuthenticationResponse = {
  identityCreated: boolean;
};

export interface TokenStore {
  set(token: string | null): void;
  get(): string | null;
}

export type Provider = {
  name: string;
  type: string;
  authorizeUrl: string;
};

export interface PasswordFlowInput {
  email: string;
  password: string;
  createIfNotExists?: boolean;
}

export interface IDTokenFlowInput {
  idToken: string;
  createIfNotExists?: boolean;
}

export interface SingleSignOnFlowInput {
  code: string;
}

type PasswordGrant = {
  grant_type: "password";
  username: string;
  password: string;
  create_if_not_exists?: boolean;
};

type TokenExchangeGrant = {
  grant_type: "token_exchange";
  subject_token: string;
  create_if_not_exists?: boolean;
};

type AuthorizationCodeGrant = {
  grant_type: "authorization_code";
  code: string;
};

type RefreshGrant = {
  grant_type: "refresh_token";
  refresh_token: string;
};

export type TokenRequest =
  | PasswordGrant
  | TokenExchangeGrant
  | AuthorizationCodeGrant
  | RefreshGrant;

export type SortDirection = "asc" | "desc" | "ASC" | "DESC";

type PageInfo = {
  count: number;
  endCursor: string;
  hasNextPage: boolean;
  startCursor: string;
  totalCount: number;
};

type FileResponseObject = {
  filename: string;
  contentType: string;
  size: number;
  url: string;
};

type unit =
  | "year"
  | "years"
  | "month"
  | "months"
  | "day"
  | "days"
  | "hour"
  | "hours"
  | "minute"
  | "minutes"
  | "second"
  | "seconds";
type direction = "next" | "last";
type completed = "complete";
type value = number;

type RelativeDateString =
  | "now"
  | "today"
  | "tomorrow"
  | "yesterday"
  | `this ${unit}`
  | `${direction} ${unit}`
  | `${direction} ${value} ${unit}`
  | `${direction} ${value} ${completed} ${unit}`;

type dateDuration =
  | `${number}Y${number}M${number}D` // Example: 1Y2M10D
  | `${number}Y${number}M` // Example: 1Y2M
  | `${number}Y${number}D` // Example: 1Y10D
  | `${number}M${number}D`; // Example: 10M2D

type timeDuration =
  | `${number}H${number}M${number}S` // Example: T2H30M
  | `${number}H${number}M` // Example: T2H30M
  | `${number}H${number}S` // Example: T2H30S
  | `${number}M` // Example: T30M
  | `${number}S`; // Example: T30S

export type DurationString =
  | `P${dateDuration}T${timeDuration}`
  | `P${dateDuration}`
  | `PT${timeDuration}`;
