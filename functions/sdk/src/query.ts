import { DatabasePool, TaggedTemplateLiteralInvocation } from 'slonik';
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
  BuiltInFields
} from './types';

export class ChainableQuery<T> {
  private readonly tableName: string;
  private readonly conditions : Conditions<T>[];
  private readonly pool: DatabasePool;

  constructor({ tableName, pool, conditions }: ChainedQueryOpts<T>) {
    this.tableName = tableName;
    this.conditions = conditions;
    this.pool = pool;
  }

  // orWhere can be used to chain additional conditions to a pre-existent set of conditions
  orWhere = (conditions: Conditions<T>) : ChainableQuery<T> => {
    this.appendConditions(conditions);

    return this;
  };

  // All causes a query to be executed, and all of the results matching the conditions
  // will be returned
  all = async () : Promise<T[]> => {
    const sql = buildSelectStatement<T>(this.tableName, this.conditions);

    const result = await this.pool.connect(async (connection) => {
      return connection.query(sql);
    });

    return result.rows as T[];
  };

  // findOne returns one record even if multiple are returned in the result set
  findOne = async () : Promise<T> => {
    const sql = buildSelectStatement<T>(this.tableName, this.conditions);

    const result = await this.pool.connect(async (connection) => {
      return connection.query(sql);
    });

    return result.rows[0];
  };

  // Returns the SQL string representing the query
  sql = ({ asAst }: SqlOptions) : string | TaggedTemplateLiteralInvocation<T> => {
    if (asAst) {
      return buildSelectStatement(this.tableName, this.conditions);
    }

    return buildSelectStatement(this.tableName, this.conditions).sql;
  };

  private appendConditions(conditions: Conditions<T>) : void {
    this.conditions.push(conditions);
  }
}

export default class Query<T> {
  private readonly tableName: string;
  private readonly conditions : Conditions<T>[];
  private readonly pool: DatabasePool;

  constructor({ tableName, pool }: QueryOpts) {
    this.tableName = tableName;
    this.conditions = [];
    this.pool = pool;
  }

  create = async (inputs: Partial<T>) : Promise<T> => {
    const now = new Date();
    const ksuid = await KSUID.random(now);
    const builtIns : BuiltInFields = {
      id: ksuid.string,
      createdAt: now,
      updatedAt: now
    };

    const query = buildCreateStatement(this.tableName, inputs, builtIns);

    const result = await this.pool.connect(async (connection) => {
      return connection.query(query);
    });

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
      conditions: [conditions]
    });
  };

  delete = async (id: string) : Promise<boolean> => {
    const query = buildDeleteStatement(this.tableName, id);

    const result = await this.pool.connect(async (connection) => {
      return connection.query(query);
    });

    return result.rowCount === 1;
  };

  findOne = async (conditions: Conditions<T>) : Promise<T> => {
    const query = buildSelectStatement<T>(this.tableName, [conditions]);

    const result = await this.pool.connect(async (connection) => {
      return connection.query(query);
    });

    return result.rows[0];
  };

  update = async (id: string, inputs: Input<T>) : Promise<T> => {
    // todo type below correctly.
    const query = buildUpdateStatement(this.tableName, id, inputs as any);

    await this.pool.connect(async (connection) => {
      return connection.query(query);
    });

    // todo: return whole object
    return inputs as T;
  };

  all = async () : Promise<T[]> => {
    const sql = buildSelectStatement(this.tableName, this.conditions);

    const result = await this.pool.connect(async (connection) => {
      return connection.query(sql);
    });

    return result.rows as T[];
  };
}

