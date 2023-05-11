const { applyWhereConditions } = require("./applyWhereConditions");
const { applyJoins } = require("./applyJoins");
const { camelCaseObject, upperCamelCase } = require("./casing");
const tracing = require("./tracing");

class QueryBuilder {
  /**
   * @param {string} tableName
   * @param {import("./QueryContext").QueryContext} context
   * @param {import("kysely").Kysely} db
   */
  constructor(tableName, context, db) {
    this._tableName = tableName;
    this._context = context;
    this._db = db;
  }

  where(where) {
    const context = this._context.clone();

    let builder = applyJoins(context, this._db, where);
    builder = applyWhereConditions(context, builder, where);

    return new QueryBuilder(this._tableName, context, builder);
  }

  orWhere(where) {
    const context = this._context.clone();

    let builder = applyJoins(context, this._db, where);

    builder = builder.orWhere((qb) => {
      return applyWhereConditions(context, qb, where);
    });

    return new QueryBuilder(this._tableName, context, builder);
  }

  async findMany() {
    const spanName = `Database ${upperCamelCase(this._tableName)}.findMany`;
    return tracing.withSpan(spanName, async (span) => {
      const query = this._db.orderBy("id");
      span.setAttribute("sql", query.compile().sql);
      const rows = await query.execute();
      return rows.map((x) => camelCaseObject(x));
    });
  }
}

module.exports.QueryBuilder = QueryBuilder;
