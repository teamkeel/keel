import { snakeCase } from "./casing";

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

export { applyLimit, applyOffset, applyOrderBy };
