const { snakeCase } = require("./casing");
const { sql } = require("kysely");

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

function applyWhereConditions(qb, where) {
  for (const key of Object.keys(where)) {
    const v = where[key];
    const fieldName = snakeCase(key);

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

module.exports.applyWhereConditions = applyWhereConditions;
