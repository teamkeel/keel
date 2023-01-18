const { applyWhereConditions } = require("./applyWhereConditions");
const { camelCaseObject } = require("./casing");

class QueryBuilder {
  constructor(db) {
    this._db = db;
  }

  where(conditions) {
    const q = this._db.where((qb) => {
      return applyWhereConditions(qb, conditions);
    });
    return new QueryBuilder(q);
  }

  orWhere(conditions) {
    const q = this._db.orWhere((qb) => {
      return applyWhereConditions(qb, conditions);
    });
    return new QueryBuilder(q);
  }

  async findMany() {
    const rows = await this._db.execute();
    return rows.map((x) => camelCaseObject(x));
  }
}

module.exports.QueryBuilder = QueryBuilder;
