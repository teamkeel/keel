export type IDWhereCondition = {
  equals?: string | null;
  notEquals?: string | null;
  oneOf?: string[] | null;
};

export type StringWhereCondition = {
  startsWith?: string | null;
  endsWith?: string | null;
  oneOf?: string[] | null;
  contains?: string | null;
  equals?: string | null;
  notEquals?: string | null;
};

export type BooleanWhereCondition = {
  equals?: boolean | null;
  notEquals?: boolean | null;
};

export type NumberWhereCondition = {
  greaterThan?: number | null;
  greaterThanOrEquals?: number | null;
  lessThan?: number | null;
  lessThanOrEquals?: number | null;
  equals?: number | null;
  notEquals?: number | null;
};

export type DateWhereCondition = {
  equals?: Date | string | null;
  before?: Date | string | null;
  onOrBefore?: Date | string | null;
  after?: Date | string | null;
  onOrAfter?: Date | string | null;
};

export type DateQueryInput = {
  equals?: string | null;
  before?: string | null;
  onOrBefore?: string | null;
  after?: string | null;
  onOrAfter?: string | null;
};

export type TimestampQueryInput = {
  before: string | null;
  after: string | null;
};

export type StringArrayWhereCondition = {
  equals?: string[] | null;
  notEquals?: string[] | null;
  any?: StringArrayQueryWhereCondition | null;
  all?: StringArrayQueryWhereCondition | null;
};

export type StringArrayQueryWhereCondition = {
  equals?: string | null;
  notEquals?: string | null;
};

export type NumberArrayWhereCondition = {
  equals?: number[] | null;
  notEquals?: number[] | null;
  any?: NumberArrayQueryWhereCondition | null;
  all?: NumberArrayQueryWhereCondition | null;
};

export type NumberArrayQueryWhereCondition = {
  greaterThan?: number | null;
  greaterThanOrEquals?: number | null;
  lessThan?: number | null;
  lessThanOrEquals?: number | null;
  equals?: number | null;
  notEquals?: number | null;
};

export type BooleanArrayWhereCondition = {
  equals?: boolean[] | null;
  notEquals?: boolean[] | null;
  any?: BooleanArrayQueryWhereCondition | null;
  all?: BooleanArrayQueryWhereCondition | null;
};

export type BooleanArrayQueryWhereCondition = {
  equals?: boolean | null;
  notEquals?: boolean | null;
};

export type DateArrayWhereCondition = {
  equals?: Date[] | null;
  notEquals?: Date[] | null;
  any?: DateArrayQueryWhereCondition | null;
  all?: DateArrayQueryWhereCondition | null;
};

export type DateArrayQueryWhereCondition = {
  greaterThan?: Date | null;
  greaterThanOrEquals?: Date | null;
  lessThan?: Date | null;
  lessThanOrEquals?: number | null;
  equals?: Date | null;
  notEquals?: Date | null;
};

export type ContextAPI = {
  headers: RequestHeaders;
  response: Response;
  isAuthenticated: boolean;
  now(): Date;
};

export type Response = {
  headers: Headers;
  status?: number;
};

export type PageInfo = {
  startCursor: string;
  endCursor: string;
  totalCount: number;
  hasNextPage: boolean;
  count: number;
};

export type FileRecord = {
  key: string;
  filename: string;
  contentType: string;
  size: number;
};

// Request headers cannot be mutated, so remove any methods that mutate
export type RequestHeaders = Omit<Headers, "append" | "delete" | "set">;

export declare class Permissions {
  constructor();

  // allow() can be used to explicitly permit access to an action
  allow(): void;

  // deny() can be used to explicitly deny access to an action
  deny(): never;
}

declare class NotFoundError extends Error {}
declare class BadRequestError extends Error {}
declare class UnknownError extends Error {}

export type Errors = {
  /**
   * Returns a 404 HTTP status with an optional message.
   * This error indicates that the requested resource could not be found.
   */
  NotFound: typeof NotFoundError;
  /**
   * Returns a 400 HTTP status with an optional message.
   * This error indicates that the request made by the client is invalid or malformed.
   */
  BadRequest: typeof BadRequestError;
  /**
   * Returns a 500 HTTP status with an optional message.
   * This error indicates that an unexpected condition was encountered, preventing the server from fulfilling the request.
   */
  Unknown: typeof UnknownError;
};
