export type SqlQueryParts = SqlQueryPart[];

export type SqlQueryPart = RawSql | SqlInput;

export type RawSql = {
  type: "sql";
  value: string;
};

export type SqlInput = {
  type: "input";
  value: any;
};

export function rawSql(sql: string): SqlQueryPart {
  return {
    type: "sql",
    value: sql,
  };
}

export function sqlIdentifier(
  identifier: string,
  identifierField: string | null = null
): SqlQueryPart {
  return {
    type: "sql",
    value: `"${identifier}"` + (identifierField ? `."${identifierField}"` : ""),
  };
}

export function sqlInput(input: any): SqlQueryPart {
  return {
    type: "input",
    value: input,
  };
}

export function sqlInputArray(inputs: any[]): SqlQueryPart[] {
  return inputs.map(sqlInput);
}

export function sqlAddSeparator(
  parts: SqlQueryPart[],
  separator: SqlQueryPart
): SqlQueryPart[] {
  if (parts.length == 0) {
    return parts;
  } else {
    const result = [parts[0]];
    for (let i = 1; i < parts.length; i++) {
      result.push(separator, parts[i]);
    }
    return result;
  }
}

export function sqlAddSeparatorAndFlatten(
  parts: SqlQueryPart[][],
  separator: SqlQueryPart
): SqlQueryPart[] {
  if (parts.length == 0) {
    return [];
  } else {
    const result = [...parts[0]];
    for (let i = 1; i < parts.length; i++) {
      result.push(separator, ...parts[i]);
    }
    return result;
  }
}
