const { applyWhereConditions } = require("./applyWhereConditions");
const {
  applyLimit,
  applyOffset,
  applyOrderBy,
} = require("./applyAdditionalQueryConstraints");
const { applyJoins } = require("./applyJoins");
const { camelCaseObject, snakeCaseObject } = require("./casing");
const { useDatabase } = require("./database");
const { QueryContext } = require("./QueryContext");
const tracing = require("./tracing");
const { DatabaseError } = require("./errors");

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

  async update(values) {
    const name = tracing.spanNameForModelAPI(this._modelName, "update");
    const db = useDatabase();

    return tracing.withSpan(name, async (span) => {
      // we build a sub-query to add to the WHERE id IN (XXX) containing all of the
      // wheres added in previous .where() chains.
      const sub = this._db.clearSelect().select("id");

      const query = db
        .updateTable(this._tableName)
        .set(snakeCaseObject(values))
        .returningAll()
        .where("id", "in", sub);

      try {
        const result = await query.execute();
        const numUpdatedRows = result.length;

        // the double (==) is important because we are comparing bigint to int
        if (numUpdatedRows == 0) {
          return null;
        }

        if (numUpdatedRows > 1) {
          throw new DatabaseError(
            new Error(
              "more than one row matched update constraints - only unique fields should be used when updating."
            )
          );
        }

        return camelCaseObject(result[0]);
      } catch (e) {
        throw new DatabaseError(e);
      }
    });
  }

  async delete() {
    const name = tracing.spanNameForModelAPI(this._modelName, "delete");
    const db = useDatabase();

    return tracing.withSpan(name, async (span) => {
      // the original query selects the distinct id + the model.* so we need to clear
      const sub = this._db.clearSelect().select("id");
      let builder = db.deleteFrom(this._tableName).where("id", "in", sub);

      const query = builder.returning(["id"]);

      // final query looks something like:
      // delete from "person" where "id" in (select distinct on ("person"."id") "id" from "person" where "person"."id" = $1) returning "id"

      span.setAttribute("sql", query.compile().sql);

      try {
        const row = await query.executeTakeFirstOrThrow();
        return row.id;
      } catch (e) {
        throw new DatabaseError(e);
      }
    });
  }

  async findOne() {
    const name = tracing.spanNameForModelAPI(this._modelName, "findOne");
    const db = useDatabase();

    return tracing.withSpan(name, async (span) => {
      let builder = db
        .selectFrom((qb) => {
          // this._db contains all of the where constraints and joins
          // we want to include that in the sub query in the same way we
          // add all of this information into the sub query in the ModelAPI's
          // implementation of findOne
          return this._db.as(this._tableName);
        })
        .selectAll();

      span.setAttribute("sql", builder.compile().sql);

      const row = await builder.executeTakeFirstOrThrow();

      if (!row) {
        return null;
      }

      return camelCaseObject(row);
    });
  }

  // orderBy(conditions) {
  //   const context = this._context.clone();

  //   const builder = applyOrderBy(
  //     context,
  //     this._db,
  //     this._tableName,
  //     conditions
  //   );

  //   return new QueryBuilder(this._tableName, context, builder);
  // }

  // limit(limit) {
  //   const context = this._context.clone();
  //   const builder = applyLimit(context, this._db, limit);

  //   return new QueryBuilder(this._tableName, context, builder);
  // }

  // offset(offset) {
  //   const context = this._context.clone();
  //   const builder = applyOffset(context, builder, offset);

  //   return new QueryBuilder(this._tableName, context, builder);
  // }

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
