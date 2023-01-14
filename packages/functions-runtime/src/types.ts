import { JSONRPCRequest, JSONRPCResponse } from "json-rpc-2.0";

import {
  StringConstraint,
  BooleanConstraint,
  NumberConstraint,
  DateConstraint,
  EnumConstraint,
} from "./constraints";
import { Logger } from "./";
import Query from "./query";
import { QueryResolver } from "./db/resolver";

export interface QueryOpts {
  tableName: string;
  queryResolver: QueryResolver;
  logger: Logger;
}

export interface ChainedQueryOpts<T> extends QueryOpts {
  conditions: Conditions<T>[];
}

export type Constraints<T> = T extends String
  ? StringConstraint
  : T extends Boolean
  ? BooleanConstraint
  : T extends Number
  ? NumberConstraint
  : T extends Date
  ? DateConstraint
  : EnumConstraint;

export type Input<T> = Record<keyof T, unknown>;

export type Conditions<T> = Partial<{ [K in keyof T]: Constraints<T[K]> }>;

export interface BuiltInFields {
  id: string;
  createdAt: Date;
  updatedAt: Date;
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

// https://www.jsonrpc.org/specification#request_object
export type CustomFunctionRequestPayload = JSONRPCRequest;

export type CustomFunctionResponsePayload = JSONRPCResponse;

export type CustomFunction = (inputs: any, api: API) => Promise<any>;

type API = {
  [apiName: string]: Query<BuiltInFields>;
};

export type Functions = Record<string, CustomFunction>;

// Config represents the configuration values
// to be passed to the Custom Code runtime server
export interface Config {
  functions: Functions;
  api: API;
}
