import {
  DatabasePool,
  QueryResult,
  TaggedTemplateLiteralInvocation
} from 'slonik';
import KSUID from 'ksuid';
import {
  buildCreateStatement,
  buildSelectStatement,
  buildUpdateStatement,
  buildDeleteStatement
} from './queryBuilders';
import {
  Conditions,
  ChainedQueryOpts,
  SqlOptions,
  QueryOpts,
  Input,
  BuiltInFields,
  OrderClauses
} from './types';
import Logger from './logger';
import { LogLevel } from './';

export class ChainableQuery<T extends IDer> {
  private readonly tableName: string;
  private readonly conditions : Conditions<T>[];
  private orderClauses: OrderClauses<T>;
  private readonly pool: DatabasePool;
  private readonly logger: Logger;

  constructor({ tableName, pool, conditions, logger }: ChainedQueryOpts<T>) {
    this.tableName = tableName;
    this.conditions = conditions;
    this.pool = pool;
    this.logger = logger;
  }

  // orWhere can be used to chain additional conditions to a pre-existent set of conditions
  orWhere = (conditions: Conditions<T>) : ChainableQuery<T> => {
    this.appendConditions(conditions);

    return this;
  };

  // All causes a query to be executed, and all of the results matching the conditions
  // will be returned
  all = async () : Promise<T[]> => {
    const sql = buildSelectStatement<T>(this.tableName, this.conditions, this.orderClauses);

    const result = await this.execute(sql);

    return result.rows as T[];
  };

  // findOne returns one record even if multiple are returned in the result set
  findOne = async () : Promise<T> => {
    const sql = buildSelectStatement<T>(this.tableName, this.conditions);

    const result = await this.execute(sql);

    return result.rows[0];
  };

  order = (clauses: OrderClauses<T>) : ChainableQuery<T> => {
    this.orderClauses = { ...this.orderClauses, ...clauses };

    return this;
  };

  // Returns the SQL string representing the query
  sql = ({ asAst }: SqlOptions) : string | TaggedTemplateLiteralInvocation<T> => {
    if (asAst) {
      return buildSelectStatement(this.tableName, this.conditions, this.orderClauses);
    }

    return buildSelectStatement(this.tableName, this.conditions).sql;
  };

  private appendConditions(conditions: Conditions<T>) : void {
    this.conditions.push(conditions);
  }

  private execute = async (query: TaggedTemplateLiteralInvocation<T>) : Promise<QueryResult<T>> => {
    this.logger.log(logSql<T>(query), LogLevel.Debug);

    return this.pool.connect(async (connection) => {
      return connection.query(query);
    });
  };
}

interface IDer {
  id: string
}

export default class Query<T extends IDer> {
  private readonly tableName: string;
  private readonly conditions : Conditions<T>[];
  private readonly pool: DatabasePool;
  private readonly logger: Logger;

  constructor({ tableName, pool, logger }: QueryOpts) {
    this.tableName = tableName;
    this.conditions = [];
    this.pool = pool;
    this.logger = logger;
  }

  create = async (inputs: Partial<T>) : Promise<T> => {
    const now = new Date();
    const ksuid = await KSUID.random(now);
    const builtIns : BuiltInFields = {
      id: ksuid.string,
      createdAt: now.toISOString(),
      updatedAt: now.toISOString()
    };

    const values = { ...inputs, ...builtIns };

    const query = buildCreateStatement(this.tableName, values);

    const result = await this.execute(query);

    // todo: better typing here
    return {
      ...inputs,
      id: result.rows[0].id as string
    } as unknown as T;
  };

  where = (conditions: Conditions<T>) : ChainableQuery<T> => {
    // ChainableQuery has a slightly different API to Query
    // as we do not want to expose methods that should only be chained
    // at the top level e.g Query.orWhere doesnt make much sense.
    return new ChainableQuery({
      tableName: this.tableName,
      pool: this.pool,
      conditions: [conditions],
      logger: this.logger
    });
  };

  delete = async (id: string) : Promise<boolean> => {
    const query = buildDeleteStatement(this.tableName, id);

    const result = await this.execute(query);

    return result.rowCount === 1;
  };

  findOne = async (conditions: Conditions<T>) : Promise<T> => {
    const query = buildSelectStatement<T>(this.tableName, [conditions]);

    const result = await this.execute(query);

    return result.rows[0];
  };

  update = async (id: string, inputs: Input<T>) : Promise<T> => {
    // todo type below correctly.
    const query = buildUpdateStatement(this.tableName, id, inputs as any);

    await this.execute(query);

    // todo: return whole object
    return inputs as T;
  };

  all = async () : Promise<T[]> => {
    const sql = buildSelectStatement(this.tableName, this.conditions);

    const result = await this.execute(sql);

    return result.rows as T[];
  };

  private execute = (query: TaggedTemplateLiteralInvocation<T>) : Promise<QueryResult<T>> => {
    this.logger.log(logSql<T>(query), LogLevel.Debug);

    return this.pool.connect(async (connection) => {
      const result = connection.query(query);
      return result;
    });
  };
}

const logSql = <T extends IDer>(query: TaggedTemplateLiteralInvocation<T>) : string => {
  const mutatedQuery = query.values.reduce((acc, cur, idx) => {
    const newObj = Object.assign({}, acc);

    const v = typeof cur.valueOf();

    let value = '';

    switch(v) {
    case 'number':
    case 'boolean':
      value = cur.toString();
      break;
    case 'string':
      value = `'${cur}'`;
      break;
    default:
      value = `'${JSON.stringify(cur)}'`;
    }

    const newSql = newObj.sql.replace(`$${idx + 1}`, value);

    return Object.assign(newObj, { sql: newSql });
  }, query);
  
  return mutatedQuery.sql;
};