import {
  sql,
  ValueExpression,
  TaggedTemplateLiteralInvocation,
} from 'slonik';
import {
  Conditions,
  Constraints,
  OrderClauses
} from '../types';

import toSnakeCase from '../util/snakeCaser';

// StringConstraint
const ENDS_WITH = 'endsWith';
const CONTAINS = 'contains';
const STARTS_WITH = 'startsWith';
const ONE_OF = 'oneOf';

// NumberConstraint
const GREATER_THAN = 'greaterThan';
const LESS_THAN = 'lessThan';
const GREATER_THAN_OR_EQUAL_TO = 'greaterThanOrEqualTo';
const LESS_THAN_OR_EQUAL_TO = 'lessThanOrEqualTo';

// EqualityConstraint
const NOT_EQUAL = 'notEqual';
const EQUAL = 'equal';

// DateConstraint
const ON_OR_BEFORE = 'onOrBefore';
const BEFORE = 'before';
const AFTER = 'after';
const ON_OR_AFTER = 'onOrAfter';

export const buildSelectStatement = <T>(tableName: string, conditions: Conditions<T>[], order?: OrderClauses<T>, limit?: number) : TaggedTemplateLiteralInvocation<T> => {
  const ands : ValueExpression[] = [];
  const hasConditions = conditions.length > 0;
  const hasOrder = Object.keys(order || {}).length > 0;
  let query = sql`SELECT * FROM ${sql.identifier([toSnakeCase(tableName)])}`;

  if (hasConditions) {
    conditions.forEach((condition) => {
      const ors : ValueExpression[] = [];
  
      Object.entries(condition).forEach(([field, constraints]) => {
        const isComplex = isComplexConstraint(constraints);
        const fullyQualifiedField = sql.identifier([toSnakeCase(tableName), toSnakeCase(field)]);
  
        if (isComplex) {
          Object.entries(constraints).forEach(([operation, value]) => {
            switch(operation) {
            case STARTS_WITH:
              // % is part of the parameter value, so needs to be interpolated
              // instead of placed in the main body of the sql:
              // https://github.com/brianc/node-postgres/issues/503#issuecomment-32055380
              ors.push(sql`${fullyQualifiedField} ILIKE ${`${value}%`}`);
              break;
            case ENDS_WITH:
              ors.push(sql`${fullyQualifiedField} ILIKE ${`%${value}`}`);
              break;
            case CONTAINS:
              ors.push(sql`${fullyQualifiedField} ILIKE ${`%${value}%`}`);
              break;
            case ONE_OF:
              // todo: join with correct type
              if (Array.isArray(value) && value.length > 0) {
                ors.push(sql`${fullyQualifiedField} IN (${sql.join(value, sql`,`)})`);
              }
              break;
            case GREATER_THAN:
              ors.push(sql`${fullyQualifiedField} > ${value}`);
              break;
            case LESS_THAN:
              ors.push(sql`${fullyQualifiedField} < ${value}`);
              break;
            case LESS_THAN_OR_EQUAL_TO:
              ors.push(sql`${fullyQualifiedField} <= ${value}`);
              break;
            case GREATER_THAN_OR_EQUAL_TO:
              ors.push(sql`${fullyQualifiedField} >= ${value}`);
              break;
            case NOT_EQUAL:
              ors.push(sql`${fullyQualifiedField} != ${value}`);
              break;
            case EQUAL:
              ors.push(sql`${fullyQualifiedField} = ${value}`);
              break;
            case BEFORE:
              ors.push(sql`${fullyQualifiedField} < ${dateParam(value)}`);
              break;
            case AFTER:
              ors.push(sql`${fullyQualifiedField} > ${dateParam(value)}`);
              break;
            case ON_OR_AFTER:
              ors.push(sql`${fullyQualifiedField} >= ${dateParam(value)}`);
              break;
            case ON_OR_BEFORE:
              ors.push(sql`${fullyQualifiedField} <= ${dateParam(value)}`);
              break;
            default:
              throw new Error('Unrecognised constraint type');
            }
          });
        } else {
          ors.push(sql`${fullyQualifiedField} = ${constraints as ValueExpression}`);
        }
      });
  
      const s = sql.join(ors, sql` AND `);
  
      // group with ()
      const grouping = sql`(${s})`;
  
      ands.push(grouping);
    });

    const whereToken = sql.join(ands, sql` OR `);

    const limitToken = limit ? sql` LIMIT ${limit}` : sql``;
  
    query = sql`${query} WHERE ${whereToken}${limitToken}`;
  }

  if (hasOrder) {
    const orderClauses = Object.entries(order).map(([key, value]) => `${key} ${value}`);
    const orderBy = sql.join(orderClauses, sql`,`);
    query = sql`${query} ORDER BY ${orderBy}`;
  }

  return query; 
};

const isComplexConstraint = (constraint: Constraints): boolean => {
  return constraint instanceof Object && constraint.constructor === Object;
};

export const buildCreateStatement = <T>(tableName: string, inputs: Partial<T>) : TaggedTemplateLiteralInvocation<T> => {
  return sql`
    INSERT INTO ${sql.identifier([toSnakeCase(tableName)])} (${sql.join(Object.keys(inputs).map(toSnakeCase).map(f => sql.identifier([f])), sql`, `)})
    VALUES (${sql.join(Object.values(inputs), sql`, `)})
    RETURNING id`;
};

export const buildUpdateStatement = <T>(tableName: string, id: string, inputs: Partial<T>) : TaggedTemplateLiteralInvocation<T> => {
  const values = Object.entries(inputs).map(([key, value]) => {
    return sql`${toSnakeCase(key)} = ${value as any}`;
  });

  const query = sql`UPDATE ${sql.identifier([toSnakeCase(tableName)])} SET ${sql.join(values, sql`,`)} WHERE id = ${id}`;

  return query;
};

export const buildDeleteStatement = <T>(tableName: string, id: string) : TaggedTemplateLiteralInvocation<T> => {
  const query = sql`DELETE FROM ${sql.identifier([toSnakeCase(tableName)])} WHERE id = ${id} RETURNING id`;

  return query;
};

const dateParam = (d: Date) => {
  return d.toISOString();
};