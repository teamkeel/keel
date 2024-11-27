const { sql, Kysely } = require("kysely");
const { snakeCase } = require("./casing");
const { TimePeriod } = require("./TimePeriod");

const opMapping = {
  startsWith: { op: "like", value: (v) => `${v}%` },
  endsWith: { op: "like", value: (v) => `%${v}` },
  contains: { op: "like", value: (v) => `%${v}%` },
  oneOf: { op: "=", value: (v) => sql`ANY(${v})` },
  greaterThan: { op: ">" },
  greaterThanOrEquals: { op: ">=" },
  lessThan: { op: "<" },
  lessThanOrEquals: { op: "<=" },
  before: { op: "<" },
  onOrBefore: { op: "<=" },
  after: { op: ">" },
  onOrAfter: { op: ">=" },
  equals: { op: sql`is not distinct from` },
  notEquals: { op: sql`is distinct from` },
  equalsRelative: {
    op: sql`BETWEEN`,
    value: (v) =>
      sql`${sql.raw(
        TimePeriod.fromExpression(v).periodStartSQL()
      )} AND ${sql.raw(TimePeriod.fromExpression(v).periodEndSQL())}`,
  },
  beforeRelative: {
    op: "<",
    value: (v) =>
      sql`${sql.raw(TimePeriod.fromExpression(v).periodStartSQL())}`,
  },
  afterRelative: {
    op: ">=",
    value: (v) => sql`${sql.raw(TimePeriod.fromExpression(v).periodEndSQL())}`,
  },
  any: {
    isArrayQuery: true,
    greaterThan: { op: ">" },
    greaterThanOrEquals: { op: ">=" },
    lessThan: { op: "<" },
    lessThanOrEquals: { op: "<=" },
    before: { op: "<" },
    onOrBefore: { op: "<=" },
    after: { op: ">" },
    onOrAfter: { op: ">=" },
    equals: { op: "=" },
    notEquals: { op: "=", value: (v) => sql`NOT ${v}` },
  },
  all: {
    isArrayQuery: true,
    greaterThan: { op: ">" },
    greaterThanOrEquals: { op: ">=" },
    lessThan: { op: "<" },
    lessThanOrEquals: { op: "<=" },
    before: { op: "<" },
    onOrBefore: { op: "<=" },
    after: { op: ">" },
    onOrAfter: { op: ">=" },
    equals: { op: "=" },
    notEquals: { op: "=", value: (v) => sql`NOT ${v}` },
  },
};

/**
 * Applies the given where conditions to the provided Kysely
 * instance and returns the resulting new Kysely instance.
 * @param {import("./QueryContext").QueryContext} context
 * @param {import("kysely").Kysely} qb
 * @param {Object} where
 * @returns {import("kysely").Kysely}
 */
function applyWhereConditions(context, qb, where = {}) {
  const conf = context.tableConfig();

  for (const key of Object.keys(where)) {
    const v = where[key];

    // Handle nested where conditions e.g. using a join table
    if (conf && conf[key]) {
      const rel = conf[key];
      context.withJoin(rel.referencesTable, () => {
        qb = applyWhereConditions(context, qb, v);
      });
      continue;
    }

    const fieldName = `${context.tableAlias()}.${snakeCase(key)}`;

    if (Object.prototype.toString.call(v) !== "[object Object]") {
      qb = qb.where(fieldName, sql`is not distinct from`, sql`${v}`);
      continue;
    }

    for (const op of Object.keys(v)) {
      const mapping = opMapping[op];
      if (!mapping) {
        throw new Error(`invalid where condition: ${op}`);
      }

      if (mapping.isArrayQuery) {
        for (const arrayOp of Object.keys(v[op])) {
          qb = qb.where(
            mapping[arrayOp].value
              ? mapping[arrayOp].value(v[op][arrayOp])
              : sql`${v[op][arrayOp]}`,
            mapping[arrayOp].op,
            sql`${sql(op)}(${sql.ref(fieldName)})`
          );
        }
      } else {
        qb = qb.where(
          fieldName,
          mapping.op,
          mapping.value ? mapping.value(v[op]) : sql`${v[op]}`
        );
      }
    }
  }

  return qb;
}

module.exports = {
  applyWhereConditions,
};
