import {
  StringConstraint,
  BooleanConstraint,
  NumberConstraint
} from './constraints';
import { DatabasePool } from 'slonik';
import { Logger } from './';

export interface QueryOpts {
  tableName: string;
  pool: DatabasePool;
  logger: Logger
}

export interface ChainedQueryOpts<T> extends QueryOpts {
  conditions: Conditions<T>[];
}

export interface SqlOptions {
  asAst: boolean
}

export type Constraints = StringConstraint | BooleanConstraint | NumberConstraint

export type Input<T> = Record<keyof T, unknown>

export type Conditions<T> = Partial<Record<keyof T, Constraints>>

export interface BuiltInFields {
  id: string
  createdAt: string
  updatedAt: string
}

export type OrderDirection = 'ASC' | 'DESC'
export type OrderClauses<T> = Record<keyof T, OrderDirection>
