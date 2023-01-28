const { applyWhereConditions } = require("./applyWhereConditions");
const { applyJoins } = require("./applyJoins");
const { camelCaseObject } = require("./casing");

class QueryBuilder {
  /**
   * @param {import("./QueryContext").QueryContext} context
   * @param {import("kysely").Kysely} db
   */
  constructor(context, db) {
    this._context = context;
    this._db = db;
  }

  where(where) {
    const context = this._context.clone();

    let builder = applyJoins(context, this._db, where);
    builder = applyWhereConditions(context, builder, where);

    return new QueryBuilder(context, builder);
  }

  orWhere(where) {
    const context = this._context.clone();

    let builder = applyJoins(context, this._db, where);

    builder = builder.orWhere((qb) => {
      return applyWhereConditions(context, qb, where);
    });

    return new QueryBuilder(context, builder);
  }

  async findMany() {
    const rows = await this._db.orderBy("id").execute();
    return rows.map((x) => camelCaseObject(x));
  }
}

module.exports.QueryBuilder = QueryBuilder;
