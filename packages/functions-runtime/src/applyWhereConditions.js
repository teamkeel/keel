const { sql } = require("kysely");
const { snakeCase } = require("./casing");

const opMapping = {
  startsWith: { op: "like", value: (v) => `${v}%` },
  endsWith: { op: "like", value: (v) => `%${v}` },
  contains: { op: "like", value: (v) => `%${v}%` },
  oneOf: { op: "in" },
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
};

/**
 * Applies the given where conditions to the provided Kysely
 * instance and returns the resulting new Kysely instance.
 * @param {import("./QueryContext").QueryContext} context
 * @param {import("kysely").Kysely} qb
 * @param {Object} where
 * @returns {import("kysely").Kysely}
 */
function applyWhereConditions(context, qb, where) {
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
      qb = qb.where(fieldName, "=", v);
      continue;
    }

    for (const op of Object.keys(v)) {
      const mapping = opMapping[op];
      if (!mapping) {
        throw new Error(`invalid where condition: ${op}`);
      }

      qb = qb.where(
        fieldName,
        mapping.op,
        mapping.value ? mapping.value(v[op]) : v[op]
      );
    }
  }

  return qb;
}

module.exports = {
  applyWhereConditions,
};
