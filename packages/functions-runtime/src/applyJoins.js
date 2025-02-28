const { snakeCase } = require("./casing");

/**
 * Adds the joins required by the where conditions to the given
 * Kysely instance and returns the resulting new Kysely instance.
 * @param {import("./QueryContext").QueryContext} context
 * @param {import("kysely").Kysely} qb
 * @param {Object} where
 * @returns {import("kysely").Kysely}
 */
function applyJoins(context, qb, where) {
  const conf = context.tableConfig();
  if (!conf) {
    return qb;
  }

  const srcTable = context.tableAlias();

  for (const key of Object.keys(where)) {
    const rel = conf[snakeCase(key)];
    if (!rel) {
      continue;
    }

    const targetTable = rel.referencesTable;

    if (context.hasJoin(targetTable)) {
      continue;
    }

    context.withJoin(targetTable, () => {
      switch (rel.relationshipType) {
        case "hasMany":
          // For hasMany the primary key is on the source table
          // and the foreign key is on the target table
          qb = qb.innerJoin(
            `${targetTable} as ${context.tableAlias()}`,
            `${srcTable}.id`,
            `${context.tableAlias()}.${rel.foreignKey}`
          );
          break;

        case "belongsTo":
          // For belongsTo the primary key is on the target table
          // and the foreign key is on the source table
          qb = qb.innerJoin(
            `${targetTable} as ${context.tableAlias()}`,
            `${srcTable}.${rel.foreignKey}`,
            `${context.tableAlias()}.id`
          );
          break;
        default:
          throw new Error(`unknown relationshipType: ${rel.relationshipType}`);
      }

      // Keep traversing through the where conditions to see if
      // more joins need to be applied
      qb = applyJoins(context, qb, where[key]);
    });
  }

  return qb;
}

module.exports = {
  applyJoins,
};
