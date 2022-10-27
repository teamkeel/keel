import pg from "pg";
import { SqlQueryParts } from "./query";

export interface QueryResolver {
  runQuery(query: SqlQueryParts): Promise<QueryResult>;
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
    default:
      throw Error("unexpected DB_CONN_TYPE: " + dbConnType);
  }
}

export class PgQueryResolver implements QueryResolver {
  private readonly pool: pg.Pool;
  constructor(config: { connectionString: string }) {
    this.pool = new pg.Pool({ connectionString: config.connectionString });
  }

  runQuery(query: SqlQueryParts): Promise<QueryResult> {
    return this.pool.query(this.toQuery(query));
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
