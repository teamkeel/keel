import {
  Conditions,
  Constraints,
  OrderClauses,
  OrderDirection,
} from "../types";

import toSnakeCase from "../util/snakeCaser";
import {
  rawSql,
  sqlAddSeparator,
  sqlAddSeparatorAndFlatten,
  sqlIdentifier,
  sqlInput,
  sqlInputArray,
  SqlQueryParts,
} from "../db/query";

// StringConstraint
const ENDS_WITH = "endsWith";
const CONTAINS = "contains";
const STARTS_WITH = "startsWith";
const ONE_OF = "oneOf";

// NumberConstraint
const GREATER_THAN = "greaterThan";
const LESS_THAN = "lessThan";
const GREATER_THAN_OR_EQUAL_TO = "greaterThanOrEqualTo";
const LESS_THAN_OR_EQUAL_TO = "lessThanOrEqualTo";

// EqualityConstraint
const NOT_EQUAL = "notEqual";
const EQUAL = "equal";

// DateConstraint
const ON_OR_BEFORE = "onOrBefore";
const BEFORE = "before";
const AFTER = "after";
const ON_OR_AFTER = "onOrAfter";

export const buildSelectStatement = <T>(
  tableName: string,
  conditions: Conditions<T>[],
  order?: OrderClauses<T>,
  limit?: number
): SqlQueryParts => {
  const ands: SqlQueryParts[] = [];
  const hasConditions = conditions.length > 0;
  const hasOrder = Object.keys(order || {}).length > 0;
  let query: SqlQueryParts = [
    rawSql("SELECT * FROM"),
    sqlIdentifier(toSnakeCase(tableName)),
  ];

  if (hasConditions) {
    conditions.forEach((condition) => {
      const ors: SqlQueryParts[] = [];

      Object.entries(condition).forEach(([field, constraints]) => {
        const isComplex = isComplexConstraint(constraints);
        const fullyQualifiedField = sqlIdentifier(
          toSnakeCase(tableName),
          toSnakeCase(field)
        );

        if (isComplex) {
          Object.entries(constraints).forEach(([operation, value]) => {
            switch (operation) {
              case STARTS_WITH:
                // % is part of the parameter value, so needs to be interpolated
                // instead of placed in the main body of the sql
                ors.push([
                  fullyQualifiedField,
                  rawSql("ILIKE"),
                  sqlInput(`${value}%`),
                ]);
                break;
              case ENDS_WITH:
                ors.push([
                  fullyQualifiedField,
                  rawSql("ILIKE"),
                  sqlInput(`%${value}`),
                ]);
                break;
              case CONTAINS:
                ors.push([
                  fullyQualifiedField,
                  rawSql("ILIKE"),
                  sqlInput(`%${value}%`),
                ]);
                break;
              case ONE_OF:
                // todo: join with correct type
                if (Array.isArray(value) && value.length > 0) {
                  ors.push([
                    fullyQualifiedField,
                    rawSql("IN ("),
                    ...sqlAddSeparator(sqlInputArray(value), rawSql(`,`)),
                    rawSql(")"),
                  ]);
                }
                break;
              case GREATER_THAN:
                ors.push([fullyQualifiedField, rawSql(">"), sqlInput(value)]);
                break;
              case LESS_THAN:
                ors.push([fullyQualifiedField, rawSql("<"), sqlInput(value)]);
                break;
              case LESS_THAN_OR_EQUAL_TO:
                ors.push([fullyQualifiedField, rawSql("<="), sqlInput(value)]);
                break;
              case GREATER_THAN_OR_EQUAL_TO:
                ors.push([fullyQualifiedField, rawSql(">="), sqlInput(value)]);
                break;
              case NOT_EQUAL:
                ors.push([
                  fullyQualifiedField,
                  rawSql("IS DISTINCT FROM"),
                  sqlInput(value),
                ]);
                break;
              case EQUAL:
                ors.push([
                  fullyQualifiedField,
                  rawSql("IS NOT DISTINCT FROM"),
                  sqlInput(value),
                ]);
                break;
              case BEFORE:
                ors.push([
                  fullyQualifiedField,
                  rawSql("<"),
                  sqlInput(dateParam(value)),
                ]);
                break;
              case AFTER:
                ors.push([
                  fullyQualifiedField,
                  rawSql(">"),
                  sqlInput(dateParam(value)),
                ]);
                break;
              case ON_OR_AFTER:
                ors.push([
                  fullyQualifiedField,
                  rawSql(">="),
                  sqlInput(dateParam(value)),
                ]);
                break;
              case ON_OR_BEFORE:
                ors.push([
                  fullyQualifiedField,
                  rawSql("<="),
                  sqlInput(dateParam(value)),
                ]);
                break;
              default:
                throw new Error("Unrecognised constraint type");
            }
          });
        } else {
          ors.push([fullyQualifiedField, rawSql("="), sqlInput(constraints)]);
        }
      });

      const s = sqlAddSeparatorAndFlatten(ors, rawSql("AND"));

      // group with ()
      const grouping = [rawSql("("), ...s, rawSql(")")];

      ands.push(grouping);
    });

    const whereToken = sqlAddSeparatorAndFlatten(ands, rawSql("OR"));

    const limitToken = limit ? [rawSql("LIMIT"), sqlInput(limit)] : [];

    query = [...query, rawSql("WHERE"), ...whereToken, ...limitToken];
  }

  if (hasOrder) {
    const orderClauses = Object.entries(order).map(([key, value]) => {
      if (value === "ASC" || value === "DESC") {
        let order: OrderDirection = value;
        return [rawSql(key), rawSql(value)];
      } else {
        throw new Error("Unrecognised order value");
      }
    });
    const orderBy = sqlAddSeparatorAndFlatten(orderClauses, rawSql(","));
    query = [...query, rawSql("ORDER BY"), ...orderBy];
  }

  return query;
};

const isComplexConstraint = (constraint: Constraints): boolean => {
  return constraint instanceof Object && constraint.constructor === Object;
};

export const buildCreateStatement = <T>(
  tableName: string,
  inputs: Partial<T>
): SqlQueryParts => {
  return [
    rawSql("INSERT INTO"),
    sqlIdentifier(toSnakeCase(tableName)),
    rawSql("("),
    ...sqlAddSeparator(
      Object.keys(inputs)
        .map(toSnakeCase)
        .map((f) => sqlIdentifier(f)),
      rawSql(",")
    ),
    rawSql(") VALUES ("),
    ...sqlAddSeparator(Object.values(inputs).map(sqlInput), rawSql(",")),
    rawSql(") RETURNING id"),
  ];
};

export const buildUpdateStatement = <T>(
  tableName: string,
  id: string,
  inputs: Partial<T>
): SqlQueryParts => {
  const values = Object.entries(inputs).map(([key, value]) => {
    return [sqlIdentifier(toSnakeCase(key)), rawSql("="), sqlInput(value)];
  });

  const query = [
    rawSql("UPDATE"),
    sqlIdentifier(toSnakeCase(tableName)),
    rawSql("SET"),
    ...sqlAddSeparatorAndFlatten(values, rawSql(",")),
    rawSql("WHERE id ="),
    sqlInput(id),
  ];

  return query;
};

export const buildDeleteStatement = <T>(
  tableName: string,
  id: string
): SqlQueryParts => {
  const query = [
    rawSql("DELETE FROM"),
    sqlIdentifier(toSnakeCase(tableName)),
    rawSql("WHERE id ="),
    sqlInput(id),
    rawSql("RETURNING id"),
  ];

  return query;
};

const dateParam = (d: Date) => {
  return d;
};
