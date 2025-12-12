import { ModelAPI } from "./ModelAPI";
import { TaskAPI } from "./TaskAPI";
import { RequestHeaders } from "./RequestHeaders";
import { handleRequest } from "./handleRequest";
import { handleJob } from "./handleJob";
import { handleSubscriber } from "./handleSubscriber";
import { handleRoute } from "./handleRoute";
import { handleFlow } from "./handleFlow";
import KSUID from "ksuid";
import { useDatabase } from "./database";
import {
  Permissions,
  PERMISSION_STATE,
  checkBuiltInPermissions,
} from "./permissions";
import * as tracing from "./tracing";
import { InlineFile, File } from "./File";
import { Duration } from "./Duration";
import { ErrorPresets } from "./errors";

// Export JS files
export {
  ModelAPI,
  TaskAPI,
  handleRequest,
  handleJob,
  handleSubscriber,
  handleRoute,
  KSUID,
  Permissions,
  PERMISSION_STATE,
  checkBuiltInPermissions,
  tracing,
  ErrorPresets,
};
export function ksuid() {
  return KSUID.randomSync().string;
}

// Export TS files

export { RequestHeaders, Duration, useDatabase, File, InlineFile, handleFlow };
export * from "./flows";
export { type UIApiResponses } from "./flows/ui/index";

// *****************************
// Static Types
// *****************************

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

export type DurationWhereCondition = {
  greaterThan?: DurationString | null;
  greaterThanOrEquals?: DurationString | null;
  lessThan?: DurationString | null;
  lessThanOrEquals?: DurationString | null;
  equals?: DurationString | null;
  notEquals?: DurationString | null;
};

export type DateWhereCondition = {
  equals?: Date | string | null;
  equalsRelative?: RelativeDateString | null;
  before?: Date | string | null;
  beforeRelative?: RelativeDateString | null;
  onOrBefore?: Date | string | null;
  after?: Date | string | null;
  afterRelative?: RelativeDateString | null;
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
  equalsRelative?: RelativeDateString | null;
  beforeRelative?: RelativeDateString | null;
  afterRelative?: RelativeDateString | null;
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
  pageNumber?: number;
};

export type SortDirection = "asc" | "desc" | "ASC" | "DESC";

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

export type FunctionConfig = {
  /**
   * All DB calls within the function will be executed within a transaction.
   * The transaction is rolled back if the function throws an error.
   */
  dbTransaction?: boolean;
};

export type FuncWithConfig<T> = T & {
  config: FunctionConfig;
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

export type RelativeDateString =
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
  | `${number}M${number}D` // Example: 10M2D
  | `${number}Y` // Example: 1Y
  | `${number}M` // Example: 1M
  | `${number}D`; // Example: 2D

type timeDuration =
  | `${number}H${number}M${number}S` // Example: 2H30M
  | `${number}H${number}M` // Example: 2H30M
  | `${number}M${number}S` // Example: 2M30S
  | `${number}H${number}S` // Example: 2H30S
  | `${number}H` // Example: 2H
  | `${number}M` // Example: 30M
  | `${number}S`; // Example: 30S

export type DurationString =
  | `P${dateDuration}T${timeDuration}`
  | `P${dateDuration}`
  | `PT${timeDuration}`;

export type TaskStatus =
  | "NEW"
  | "ASSIGNED"
  | "STARTED"
  | "COMPLETED"
  | "CANCELLED"
  | "DEFERRED";

export type Task = {
  id: string;
  topic: string;
  status: TaskStatus;
  deferredUntil?: Date;
  createdAt: Date;
  updatedAt: Date;
  assignedTo?: string;
  assignedAt?: Date;
  resolvedAt?: Date;
  flowRunId?: string;
};

export type TaskCreateOptions = {
  deferredUntil?: Date;
};
