// This file doesn't contain types describing this package, rather it contains generic types
// that are used by the generated @teamkeel/sdk package.

export type IDWhereCondition = {
  equals?: string;
  oneOf?: string[];
};

export type StringWhereCondition = {
  startsWith?: string;
  endsWith?: string;
  oneOf?: string[];
  contains?: string;
  notEquals?: string;
  equals?: string;
};

export type BooleanWhereCondition = {
  equals?: boolean;
  notEquals?: boolean;
};

export type NumberWhereCondition = {
  greaterThan?: number;
  greaterThanOrEquals?: number;
  lessThan?: number;
  lessThanOrEquals?: number;
  equals?: number;
  notEquals?: number;
};

// Date database API
export type DateWhereCondition = {
  equals?: Date | string;
  before?: Date | string;
  onOrBefore?: Date | string;
  after?: Date | string;
  onOrAfter?: Date | string;
};

// Date query input
export type DateQueryInput = {
  equals?: string;
  before?: string;
  onOrBefore?: string;
  after?: string;
  onOrAfter?: string;
};

// Timestamp query input
export type TimestampQueryInput = {
  before: string;
  after: string;
};

// Ctx API
export type ContextAPI = {
  headers: RequestHeaders;
  isAuthenticated: boolean;
  now(): Date;
};

// Request headers query API
export type RequestHeaders = {
  get(name: string): string;
  has(name: string): boolean;
};
