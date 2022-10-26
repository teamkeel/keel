import KSUID from "ksuid";
import {
  buildCreateStatement,
  buildSelectStatement,
  buildUpdateStatement,
  buildDeleteStatement,
} from "./queryBuilders";
import {
  Conditions,
  ChainedQueryOpts,
  SqlOptions,
  QueryOpts,
  Input,
  BuiltInFields,
  OrderClauses,
} from "./types";
import * as ReturnTypes from "./returnTypes";
import Logger from "./logger";
import { SqlQueryParts } from "./db/query";
import { Pool, QueryResult } from "pg";

export class ChainableQuery<T extends IDer> {
  private readonly tableName: string;
  private readonly conditions: Conditions<T>[];
  private orderClauses: OrderClauses<T>;
  private readonly connectionString: string;
  private readonly logger: Logger;

  constructor({
    tableName,
    connectionString,
    conditions,
    logger,
  }: ChainedQueryOpts<T>) {
    this.tableName = tableName;
    this.conditions = conditions;
    this.connectionString = connectionString;
    this.logger = logger;
  }

  // orWhere can be used to chain additional conditions to a pre-existent set of conditions
  orWhere = (conditions: Conditions<T>): ChainableQuery<T> => {
    this.appendConditions(conditions);

    return this;
  };

  // All causes a query to be executed, and all of the results matching the conditions
  // will be returned
  all = async (): Promise<ReturnTypes.FunctionListResponse<T>> => {
    const sql = buildSelectStatement<T>(
      this.tableName,
      this.conditions,
      this.orderClauses
    );

    const result = await this.execute(sql);

    return {
      collection: result.rows as T[],
    };
  };

  // findOne returns one record even if multiple are returned in the result set
  findOne = async (): Promise<ReturnTypes.FunctionGetResponse<T>> => {
    const sql = buildSelectStatement<T>(
      this.tableName,
      this.conditions,
      undefined,
      1
    );

    const result = await this.execute(sql);

    return {
      object: result.rows[0],
      errors: [],
    };
  };

  order = (clauses: OrderClauses<T>): ChainableQuery<T> => {
    this.orderClauses = { ...this.orderClauses, ...clauses };

    return this;
  };

  private appendConditions(conditions: Conditions<T>): void {
    this.conditions.push(conditions);
  }

  private execute = async (query: SqlQueryParts): Promise<QueryResult<T>> => {
    // todo: reinstate
    // this.logger.log(logSql<T>(query), LogLevel.Debug);

    const pool = new Pool({
      connectionString: this.connectionString,
    });

    return await pool.query(toQuery(query));
  };
}

interface IDer {
  id: string;
}

export default class Query<T extends IDer> {
  private readonly tableName: string;
  private readonly conditions: Conditions<T>[];
  private readonly connectionString: string;
  private readonly logger: Logger;

  constructor({ tableName, connectionString, logger }: QueryOpts) {
    this.tableName = tableName;
    this.conditions = [];
    this.connectionString = connectionString;
    this.logger = logger;
  }

  create = async (
    inputs: Partial<T>
  ): Promise<ReturnTypes.FunctionCreateResponse<T>> => {
    const now = new Date();
    const ksuid = await KSUID.random(now);
    const builtIns: BuiltInFields = {
      id: ksuid.string,
      createdAt: now.toISOString(),
      updatedAt: now.toISOString(),
    };

    const values = { ...inputs, ...builtIns };

    const query = buildCreateStatement(this.tableName, values);

    const result = await this.execute(query);

    return {
      object: {
        ...inputs,
        id: result.rows[0].id as string,
      } as unknown as T,
      errors: [],
    };
  };

  where = (conditions: Conditions<T>): ChainableQuery<T> => {
    // ChainableQuery has a slightly different API to Query
    // as we do not want to expose methods that should only be chained
    // at the top level e.g Query.orWhere doesnt make much sense.
    return new ChainableQuery({
      tableName: this.tableName,
      connectionString: this.connectionString,
      conditions: [conditions],
      logger: this.logger,
    });
  };

  delete = async (
    id: string
  ): Promise<ReturnTypes.FunctionDeleteResponse<T>> => {
    const query = buildDeleteStatement(this.tableName, id);

    const result = await this.execute(query);

    return {
      success: result.rowCount === 1,
    };
  };

  findOne = async (
    conditions: Conditions<T>
  ): Promise<ReturnTypes.FunctionGetResponse<T>> => {
    const query = buildSelectStatement<T>(this.tableName, [conditions]);

    const result = await this.execute(query);

    // buildSelectStatement stil returns an array despite applying a LIMIT 1
    // so return the first row anyhow.
    return {
      object: result.rows[0],
      errors: [],
    };
  };

  update = async (
    id: string,
    inputs: Input<T>
  ): Promise<ReturnTypes.FunctionUpdateResponse<T>> => {
    // todo type below correctly.
    const query = buildUpdateStatement(this.tableName, id, inputs as any);

    await this.execute(query);

    return {
      object: {
        ...inputs,
        id,
      } as T,
      errors: [],
    };
  };

  all = async (): Promise<ReturnTypes.FunctionListResponse<T>> => {
    const sql = buildSelectStatement(this.tableName, this.conditions);

    const result = await this.execute(sql);

    return {
      collection: result.rows as T[],
    };
  };

  private execute = async (query: SqlQueryParts): Promise<QueryResult<T>> => {
    // todo: reinstate
    // this.logger.log(logSql<T>(query), LogLevel.Debug);

    const pool = new Pool({
      connectionString: this.connectionString,
    });

    return await pool.query(toQuery(query));
  };
}

const logSql = <T extends IDer>(query: SqlQueryParts): string => {
  return query
    .map((queryPart) => {
      switch (queryPart.type) {
        case "sql":
          return queryPart.value;
        case "input":
          let v = queryPart.value.valueOf();
          let value = "";

          switch (v) {
            case "number":
            case "boolean":
              value = queryPart.value.toString();
              break;
            case "string":
              // string covers some other types that are stringified such as date
              value = `'${queryPart.value}'`;
              break;
            default:
              value = `'${JSON.stringify(queryPart.value)}'`;
          }
          return value;
      }
    })
    .join();
};

function toQuery(query: SqlQueryParts): { text: string; values: any[] } {
  let nextInterpolationIndex = 1;
  let values = [];
  const text = query
    .map((queryPart) => {
      switch (queryPart.type) {
        case "sql":
          return queryPart.value;
        case "input":
          values.push(queryPart.value);
          return `$${nextInterpolationIndex++}`;
      }
    })
    .join(" ");
  return { text, values };
}
