import pg from "pg";
import { rawSql, SqlQueryParts } from "./query";
import {
  ArrayValue,
  ExecuteStatementCommand,
  Field,
  RDSDataClient,
  SqlParameter,
  TypeHint,
} from "@aws-sdk/client-rds-data";
import toCamelCase from "../util/camelCaser";
import { ExecuteStatementCommandInput } from "@aws-sdk/client-rds-data/dist-types/commands/ExecuteStatementCommand";

export interface QueryResolver {
  runQuery(query: SqlQueryParts): Promise<QueryResult>;
  runRawQuery(query: string): Promise<QueryResult>;
}

export interface QueryResult {
  rows: QueryResultRow[];
}

export interface QueryResultRow {
  [column: string]: any;
}

export function queryResolverFromEnv(
  env: Record<string, string | undefined>
): QueryResolver {
  const dbConnType = env["DB_CONN_TYPE"];
  switch (dbConnType) {
    case "pg":
      const dbConn = env["DB_CONN"];
      if (!dbConn) {
        throw Error("expected DB_CONN for DB_CONN_TYPE=pg");
      }
      return new PgQueryResolver({ connectionString: dbConn });
    case "dataapi":
      const region = env["DB_REGION"];
      if (!region) {
        throw Error("expected DB_REGION for DB_CONN_TYPE=dataapi");
      }
      const dbClusterResourceArn = env["DB_RESOURCE_ARN"];
      if (!dbClusterResourceArn) {
        throw Error("expected DB_RESOURCE_ARN for DB_CONN_TYPE=dataapi");
      }
      const dbCredentialsSecretArn = env["DB_SECRET_ARN"];
      if (!dbCredentialsSecretArn) {
        throw Error("expected DB_SECRET_ARN for DB_CONN_TYPE=dataapi");
      }
      const dbName = env["DB_NAME"];
      if (!dbName) {
        throw Error("expected DB_NAME for DB_CONN_TYPE=dataapi");
      }
      return new AwsRdsDataClientQueryResolver({
        region,
        dbClusterResourceArn,
        dbCredentialsSecretArn,
        dbName,
      });
    default:
      throw Error("unexpected DB_CONN_TYPE: " + dbConnType);
  }
}

export class PgQueryResolver implements QueryResolver {
  private readonly pool: pg.Pool;
  constructor(config: { connectionString: string }) {
    this.pool = new pg.Pool({ connectionString: config.connectionString });
  }

  async runRawQuery(query: string): Promise<QueryResult> {
    return this.runQuery([rawSql(query)]);
  }

  async runQuery(query: SqlQueryParts): Promise<QueryResult> {
    const result = await this.pool.query(this.toQuery(query));
    if (result.rows) {
      result.rows = result.rows.map((row) => {
        if (row && typeof row === "object") {
          const camelCasedKeysObject = {};
          for (let key of Object.keys(row)) {
            camelCasedKeysObject[toCamelCase(key)] = row[key];
          }
          return camelCasedKeysObject;
        } else {
          return row;
        }
      });
    }
    return result;
  }

  private toQuery(query: SqlQueryParts): { text: string; values: any[] } {
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
}

export class AwsRdsDataClientQueryResolver implements QueryResolver {
  private readonly client: RDSDataClient;
  private readonly dbClusterResourceArn: string;
  private readonly dbCredentialsSecretArn: string;
  private readonly dbName: string;
  constructor(config: {
    region: string;
    dbClusterResourceArn: string;
    dbCredentialsSecretArn: string;
    dbName: string;
  }) {
    this.client = new RDSDataClient({ region: config.region });
    this.dbClusterResourceArn = config.dbClusterResourceArn;
    this.dbCredentialsSecretArn = config.dbCredentialsSecretArn;
    this.dbName = config.dbName;
  }

  async runRawQuery(query: string): Promise<QueryResult> {
    return this.runQuery([rawSql(query)]);
  }

  async runQuery(query: SqlQueryParts): Promise<QueryResult> {
    const { sql, params } = this.toQuery(query);
    const input: ExecuteStatementCommandInput = {
      resourceArn: this.dbClusterResourceArn,
      secretArn: this.dbCredentialsSecretArn,
      database: this.dbName,
      sql: sql,
      parameters: params,
      includeResultMetadata: true,
    };
    const command = new ExecuteStatementCommand(input);
    const data = await this.client.send(command);
    if (!data.records) {
      return { rows: [] };
    }
    const rows = data.records.map((fieldArray) => {
      const row: QueryResultRow = {};
      for (let i = 0; i < fieldArray.length; i++) {
        const field = fieldArray[i];
        const column = data.columnMetadata[i].name;
        const typeName = data.columnMetadata[i].typeName;
        const value = Field.visit(field, {
          isNull: function (value: boolean): any {
            return null;
          },
          booleanValue: function (value: boolean): any {
            return value;
          },
          longValue: function (value: number): any {
            return value;
          },
          doubleValue: function (value: number): any {
            return value;
          },
          stringValue: function (value: string): any {
            if (typeName === "timestamp") {
              return new Date(value);
            } else {
              return value;
            }
          },
          blobValue: function (value: Uint8Array): any {
            return value;
          },
          arrayValue: function (value: ArrayValue): any {
            return value;
          },
          _: function (name: string, value: any): any {
            return value;
          },
        });
        row[column] = value;
      }
      return row;
    });
    const rowsWithObjectKeysCamelCased = rows.map((row) => {
      if (row && typeof row === "object") {
        const camelCasedKeysObject = {};
        for (let key of Object.keys(row)) {
          camelCasedKeysObject[toCamelCase(key)] = row[key];
        }
        return camelCasedKeysObject;
      } else {
        return row;
      }
    });
    return { rows: rowsWithObjectKeysCamelCased };
  }

  private toQuery(query: SqlQueryParts): {
    sql: string;
    params: SqlParameter[];
  } {
    let nextInterpolationIndex = 1;
    let params: SqlParameter[] = [];
    const sql = query
      .map((queryPart) => {
        switch (queryPart.type) {
          case "sql":
            return queryPart.value;
          case "input":
            const paramName = `param${nextInterpolationIndex++}`;
            let field: Field;
            let typeHint: TypeHint | undefined = undefined;
            if (queryPart.value == null) {
              field = { isNull: true };
            } else if (typeof queryPart.value === "number") {
              //TODO doubleValue
              field = { longValue: queryPart.value };
            } else if (typeof queryPart.value === "boolean") {
              field = { booleanValue: queryPart.value };
            } else if (queryPart.value instanceof Date) {
              field = { stringValue: this.toDataApiFormat(queryPart.value) };
              typeHint = TypeHint.TIMESTAMP;
            } else {
              field = { stringValue: queryPart.value };
            }
            params.push({
              name: paramName,
              value: field,
              typeHint: typeHint,
            });
            return `:${paramName}`;
        }
      })
      .join(" ");
    return { sql, params };
  }

  // The accepted format is YYYY-MM-DD HH:MM:SS[.FFF]
  // source: https://boto3.amazonaws.com/v1/documentation/api/latest/reference/services/rds-data.html#RDSDataService.Client.execute_statement
  private toDataApiFormat(date: Date): string {
    const iso = date.toISOString();
    const withoutZ = iso.replace(/Z$/, "");
    const withSpaceInsteadOfT = withoutZ.replace(/T/, " ");
    return withSpaceInsteadOfT;
  }
}
