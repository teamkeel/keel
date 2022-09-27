import {
  StringConstraint,
  BooleanConstraint,
  NumberConstraint,
  DateConstraint,
  EnumConstraint,
} from "./constraints";
import { Logger } from "./";

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
  | EnumConstraint;

export type Input<T> = Record<keyof T, unknown>;

export type Conditions<T> = Partial<Record<keyof T, Constraints>>;

export interface BuiltInFields {
  id: string;
  createdAt: string;
  updatedAt: string;
}

export type OrderDirection = "ASC" | "DESC";
export type OrderClauses<T> = Partial<Record<keyof T, OrderDirection>>;

// A generic Identity interface for usage in other npm packages
// without codegenerating the whole Identity interface
// We know that Identity will implement these fields
// TODO: remove once we're codegenerating this from the schema.
export interface Identity {
  id: string;
  email: string;
}
