declare module "@teamkeel/sdk/constraints" {
  export type EqualityConstraint = {
    notEqual?: string;
    equal?: string;
  };
  export type StringConstraint =
    | string
    | {
        startsWith?: string;
        endsWith?: string;
        oneOf?: string[];
        contains?: string;
      }
    | EqualityConstraint;
  export type NumberConstraint =
    | number
    | {
        greaterThan?: number;
        greaterThanOrEqualTo?: number;
        lessThan?: number;
        lessThanOrEqualTo?: number;
        equal?: number;
        notEqual?: number;
      }
    | EqualityConstraint;
  export type DateConstraint =
    | Date
    | {
        equal?: Date;
        before?: Date;
        onOrBefore?: Date;
        after?: Date;
        onOrAfter?: Date;
      };
  export type BooleanConstraint = boolean | EqualityConstraint;
  export type EnumConstraint = string | EqualityConstraint;
}
declare module "@teamkeel/sdk/index" {
  import Query, { ChainableQuery } from "@teamkeel/sdk/query";
  import * as QueryConstraints from "@teamkeel/sdk/constraints";
  import Logger, {
    ConsoleTransport,
    Level as LogLevel,
  } from "@teamkeel/sdk/logger";
  import { Identity } from "@teamkeel/sdk/types";

  //@ts-ignore
  export * from "@teamkeel/client";

  export * from "@teamkeel/sdk/returnTypes";

  export {
    Query,
    QueryConstraints,
    ChainableQuery,
    Logger,
    ConsoleTransport,
    LogLevel,
    Identity,
  };
}

declare module "@teamkeel/sdk/logger" {
  export enum Level {
    Info = "info",
    Error = "error",
    Debug = "debug",
    Warn = "warn",
    Success = "success"
  }
  type Msg = any;

  export interface Transport {
    log: (msg: Msg, level: Level, options: LoggerOptions) => void;
  }
  export interface LoggerOptions {
    transport?: Transport;
    colorize?: boolean;
    timestamps?: boolean;
  }

  export class ConsoleTransport implements Transport {
    log: (msg: Msg, level: Level, options: LoggerOptions) => void;
  }

  export default class Logger {
    private readonly options: LoggerOptions;

    constructor(opts?: LoggerOptions);

    log: (msg: Msg, level: Level) => void;
  }
}

declare module "@teamkeel/sdk/query" {
  import { TaggedTemplateLiteralInvocation } from "slonik";
  import {
    Conditions,
    ChainedQueryOpts,
    SqlOptions, 
    QueryOpts,
    Input,
    OrderClauses,
  } from "@teamkeel/sdk/types";
  import * as ReturnTypes from "@teamkeel/sdk/returnTypes";
  export class ChainableQuery<T> {
    private readonly tableName;
    private readonly conditions;
    private readonly connectionString: string;
    constructor({
      tableName,
      connectionString,
      conditions,
    }: ChainedQueryOpts<T>);
    orWhere: (conditions: Conditions<T>) => ChainableQuery<T>;
    all: () => Promise<ReturnTypes.FunctionListResponse<T>>;
    order: (clauses: OrderClauses<T>) => ChainableQuery<T>;
    findOne: () => Promise<ReturnTypes.FunctionGetResponse<T>>;
    sql: ({ asAst }: SqlOptions) => string | TaggedTemplateLiteralInvocation<T>;
    private appendConditions;
  }
  export default class Query<T> {
    private readonly tableName;
    private readonly conditions;
    private orderClauses;
    private readonly connectionString: string;

    constructor({ tableName, connectionString, logger }: QueryOpts);
    create: (
      inputs: Partial<T>
    ) => Promise<ReturnTypes.FunctionCreateResponse<T>>;
    where: (conditions: Conditions<T>) => ChainableQuery<T>;
    delete: (id: string) => Promise<ReturnTypes.FunctionDeleteResponse<T>>;
    findOne: (
      conditions: Conditions<T>
    ) => Promise<ReturnTypes.FunctionGetResponse<T>>;
    update: (
      id: string,
      inputs: Input<T>
    ) => Promise<ReturnTypes.FunctionUpdateResponse<T>>;
    all: () => Promise<ReturnTypes.FunctionListResponse<T>>;
  }
  export {};
}

declare module "@teamkeel/sdk/queryBuilders/index" {
  import { TaggedTemplateLiteralInvocation } from "slonik";
  import { Constraints } from "@teamkeel/sdk/types";
  export const buildSelectStatement: <T>(
    tableName: string,
    conditions: Partial<Record<keyof T, Constraints>>[]
  ) => TaggedTemplateLiteralInvocation<T>;
  export const buildCreateStatement: <T>(
    tableName: string,
    inputs: Partial<T>
  ) => TaggedTemplateLiteralInvocation;
  export const buildUpdateStatement: <T>(
    tableName: string,
    id: string,
    inputs: Partial<T>
  ) => TaggedTemplateLiteralInvocation<T>;
  export const buildDeleteStatement: <T>(
    tableName: string,
    id: string
  ) => TaggedTemplateLiteralInvocation<T>;
}
declare module "@teamkeel/sdk/types" {
  import {
    StringConstraint,
    BooleanConstraint,
    NumberConstraint,
    DateConstraint,
    EnumConstraint,
    EqualityConstraint
  } from "@teamkeel/sdk/constraints";
  import { Logger } from "@teamkeel/sdk";
  export interface QueryOpts {
    tableName: string;
    connectionString: string;
    logger: Logger;
  }
  export interface ChainedQueryOpts<T> extends QueryOpts {
    conditions: Conditions<T>[];
  }
  export interface SqlOptions {
    asAst: boolean;
  }
  export type Constraints =
    | StringConstraint
    | BooleanConstraint
    | NumberConstraint
    | DateConstraint
    | EnumConstraint
    | EqualityConstraint;
  export type Input<T> = Record<keyof T, unknown>;
  export type Conditions<T> = Partial<Record<keyof T, Constraints>>;
  export type OrderDirection = "asc" | "desc";
  export type OrderClauses<T> = Partial<Record<keyof T, OrderDirection>>;

  // A generic Identity interface for usage in other npm packages
  // without codegenerating the whole Identity interface
  // We know that Identity will implement these fields
  export interface Identity {
    id: string;
    email: string;
  }
}

declare module "@teamkeel/sdk/returnTypes" {
  // ValidationErrors will be returned when interacting with
  // the Query API (creating, updating entities)
  export interface ValidationError {
    field: string;
    message: string;
    code: string;
  }

  // ExecutionError represents other misc errors
  // that can occur during the execution of a custom function
  export interface ExecutionError {
    message: string;

    // todo: implement stacks
    stack: string;
  }

  export type FunctionError = ValidationError | ExecutionError;

  export interface FunctionCreateResponse<T> {
    object?: T;
    errors?: FunctionError[];
  }

  export interface FunctionGetResponse<T> {
    object?: T;
    errors?: FunctionError[];
  }

  export interface FunctionDeleteResponse<T> {
    success: boolean;
    errors?: FunctionError[];
  }

  export interface FunctionListResponse<T> {
    collection: T[];
    errors?: FunctionError[];
    // todo: add type for pagination
  }

  export interface FunctionUpdateResponse<T> {
    object?: T;
    errors?: FunctionError[];
  }

  export interface FunctionAuthenticateResponse {
    identityId?: string;
    identityCreated: boolean;
    errors?: FunctionError[];
  }
}

declare module "@teamkeel/sdk" {
  import main = require("@teamkeel/sdk/index");
  export = main;
}
