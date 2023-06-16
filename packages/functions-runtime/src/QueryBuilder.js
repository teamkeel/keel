const { applyWhereConditions } = require("./applyWhereConditions");
const {
  applyLimit,
  applyOffset,
  applyOrderBy,
} = require("./applyAdditionalQueryConstraints");
const { applyJoins } = require("./applyJoins");
const { camelCaseObject } = require("./casing");
const { useDatabase } = require("./database");
const { QueryContext } = require("./QueryContext");
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

  async findMany(params) {
    const name = tracing.spanNameForModelAPI(this._modelName, "findMany");
    const db = useDatabase();

    return tracing.withSpan(name, async (span) => {
      const context = new QueryContext([this._tableName], this._tableConfigMap);

      let builder = db
        .selectFrom((qb) => {
          // this._db contains all of the where constraints and joins
          // we want to include that in the sub query in the same way we
          // add all of this information into the sub query in the ModelAPI's
          // implementation of findMany
          return this._db.as(this._tableName);
        })
        .selectAll();

      // The only constraints added to the main query are the orderBy, limit and offset as they are performed on the "outer" set
      if (params?.limit) {
        builder = applyLimit(context, builder, params.limit);
      }

      if (params?.offset) {
        builder = applyOffset(context, builder, params.offset);
      }

      if (
        params?.orderBy !== undefined &&
        Object.keys(params?.orderBy).length > 0
      ) {
        builder = applyOrderBy(
          context,
          builder,
          this._tableName,
          params.orderBy
        );
      } else {
        builder = builder.orderBy(`${this._tableName}.id`);
      }

      const query = builder;

      span.setAttribute("sql", query.compile().sql);
      const rows = await builder.execute();
      return rows.map((x) => camelCaseObject(x));
    });
  }
}

module.exports.QueryBuilder = QueryBuilder;
