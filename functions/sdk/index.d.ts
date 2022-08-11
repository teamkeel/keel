declare module '@teamkeel/sdk/constraints' {
  export type EqualityConstraint = {
      notEqual?: string;
      equal?: string;
  };
  export type StringConstraint = string | {
      startsWith?: string;
      endsWith?: string;
      oneOf?: string[];
      contains?: string;
  } | EqualityConstraint;
  export type NumberConstraint = number | {
      greaterThan?: number;
      greaterThanOrEqualTo?: number;
      lessThan?: number;
      lessThanOrEqualTo?: number;
      equal?: number;
      notEqual?: number;
  } | EqualityConstraint;
  export type DateConstraint = Date;
  export type BooleanConstraint = EqualityConstraint;

}
declare module '@teamkeel/sdk/index' {
  import Query, { ChainableQuery } from '@teamkeel/sdk/query';
  import * as QueryConstraints from '@teamkeel/sdk/constraints';
  import Logger from '@teamkeel/sdk/logger';
  // eslint-disable-next-line @typescript-eslint/ban-ts-comment
  //@ts-ignore
  export * from '@teamkeel/client';
  export { Query, QueryConstraints, ChainableQuery, Logger };
}

declare module '@teamkeel/sdk/logger' {
  export enum Level {
    Info = 'info',
    Error = 'error',
    Debug = 'debug',
    Warn = 'warn'
  }
  type Msg = string | boolean | undefined | null | number

  export interface Transport {
    log: (msg: Msg, level: Level, options: LoggerOptions) => void
  }
  export interface LoggerOptions {
    transport: Transport
    colorize?: boolean
  }

  export class ConsoleTransport implements Transport {
    log: (msg: Msg, level: Level, options: LoggerOptions) => void;
  }

  export default class Logger {
    private readonly options : LoggerOptions;
  
    constructor(opts: LoggerOptions);
  
    log: (msg: Msg, level: Level) => void;
  }
}

declare module '@teamkeel/sdk/query' {
  import { TaggedTemplateLiteralInvocation } from 'slonik';
  import {
    Conditions,
    ChainedQueryOpts,
    SqlOptions,
    QueryOpts,
    Input
  } from '@teamkeel/sdk/types';
  export class ChainableQuery<T> {
    private readonly tableName;
    private readonly conditions;
    private readonly pool;
    constructor({ tableName, pool, conditions }: ChainedQueryOpts<T>);
    orWhere: (conditions: Conditions<T>) => ChainableQuery<T>;
    all: () => Promise<T[]>;
    findOne: () => Promise<T>;
    sql: ({ asAst }: SqlOptions) => string | TaggedTemplateLiteralInvocation<T>;
    private appendConditions;
  }
  export default class Query<T> {
    private readonly tableName;
    private readonly conditions;
    private readonly pool;
    constructor({ tableName, pool }: QueryOpts);
    create: (inputs: Partial<T>) => Promise<T>;
    where: (conditions: Conditions<T>) => ChainableQuery<T>;
    delete: (id: string) => Promise<boolean>;
    findOne: (conditions: Conditions<T>) => Promise<T>;
    update: (id: string, inputs: Input<T>) => Promise<T>;
    all: () => Promise<T[]>;
  }
  export {};

}

declare module '@teamkeel/sdk/queryBuilders/index' {
  import { TaggedTemplateLiteralInvocation } from 'slonik';
  import { Constraints } from '@teamkeel/sdk/types';
  export const buildSelectStatement: <T>(tableName: string, conditions: Partial<Record<keyof T, Constraints>>[]) => TaggedTemplateLiteralInvocation<T>;
  export const buildCreateStatement: <T>(tableName: string, inputs: Partial<T>) => TaggedTemplateLiteralInvocation;
  export const buildUpdateStatement: <T>(tableName: string, id: string, inputs: Partial<T>) => TaggedTemplateLiteralInvocation<T>;
  export const buildDeleteStatement: <T>(tableName: string, id: string) => TaggedTemplateLiteralInvocation<T>;

}
declare module '@teamkeel/sdk/types' {
  import {
    StringConstraint,
    BooleanConstraint,
    NumberConstraint
  } from '@teamkeel/sdk/constraints';
  import { DatabasePool } from 'slonik';
  export interface QueryOpts {
      tableName: string;
      pool: DatabasePool;
  }
  export interface ChainedQueryOpts<T> extends QueryOpts {
      conditions: Conditions<T>[];
  }
  export interface SqlOptions {
      asAst: boolean;
  }
  export type Constraints = StringConstraint | BooleanConstraint | NumberConstraint;
  export type Input<T> = Record<keyof T, unknown>;
  export type Conditions<T> = Partial<Record<keyof T, Constraints>>;

}
declare module '@teamkeel/sdk' {
  import main = require('@teamkeel/sdk/index');
  export = main;
}