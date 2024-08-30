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

type MimeType =
  | "application/json"
  | "application/gzip"
  | "application/pdf"
  | "application/rtf"
  | "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
  | "application/vnd.openxmlformats-officedocument.presentationml.presentation"
  | "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
  | "application/vnd.ms-excel"
  | "application/vnd.ms-powerpoint"
  | "application/msword"
  | "application/zip"
  | "application/xml"
  | "application/x-7z-compressed"
  | "application/x-tar"
  | "image/gif"
  | "image/jpeg"
  | "image/svg+xml"
  | "image/png"
  | "text/html"
  | "text/csv"
  | "text/javascript"
  | "text/plain"
  | "text/calendar"
  | (string & {});

export type InlineFileConstructor = {
  filename: string;
  contentType: MimeType;
};
export declare class InlineFile {
  constructor(input: InlineFileConstructor);
  static fromDataURL(url: string): InlineFile;
  // Reads the contents of the file as a buffer
  read(): Promise<Buffer>;
  // Write the files contents from a buffer
  write(data: Buffer): void;
  // Persists the file
  store(expires?: Date, isPublic?: boolean): Promise<StoredFile>;
  // Gets the name of the file
  get filename(): string;
  // Gets the media type of the file contents
  get contentType(): string;
  // Gets size of the file's contents in bytes
  get size(): number;
}

export declare class StoredFile extends InlineFile {
  // Gets the stored key
  get key(): string;
  // Gets size of the file's contents in bytes
  get isPublic(): boolean;
  static fromDbRecord(input: FileDbRecord): StoredFile;
  // Persists the file
  toDbRecord(): FileDbRecord;
}

export type FileDbRecord = {
  key: string;
  filename: string;
  contentType: string;
  size: number;
};

export type SortDirection = "asc" | "desc" | "ASC" | "DESC";

type PageInfo = {
  count: number;
  endCursor: string;
  hasNextPage: boolean;
  startCursor: string;
  totalCount: number;
};
