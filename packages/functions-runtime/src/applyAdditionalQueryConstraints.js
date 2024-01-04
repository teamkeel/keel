const { snakeCase } = require("change-case");

function applyLimit(context, qb, limit) {
  return qb.limit(limit);
}

function applyOffset(context, qb, offset) {
  return qb.offset(offset);
}

function applyOrderBy(context, qb, tableName, orderBy = {}) {
  Object.entries(orderBy).forEach(([key, sortOrder]) => {
    qb = qb.orderBy(`${tableName}.${snakeCase(key)}`, sortOrder.toLowerCase());
  });
  return qb;
}

module.exports = {
  applyLimit,
  applyOffset,
  applyOrderBy,
};
