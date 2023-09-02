const { useDatabase } = require("./database");
const { QueryBuilder } = require("./QueryBuilder");
const { QueryContext } = require("./QueryContext");
const { applyWhereConditions } = require("./applyWhereConditions");
const { applyJoins } = require("./applyJoins");
const { auditDataInstance } = require("./tryExecuteFunction.js");
const { sql } = require("kysely");

const {
  applyLimit,
  applyOffset,
  applyOrderBy,
} = require("./applyAdditionalQueryConstraints");
const {
  camelCaseObject,
  snakeCaseObject,
  upperCamelCase,
} = require("./casing");
const tracing = require("./tracing");
const { DatabaseError } = require("./errors");

/**
 * RelationshipConfig is a simple representation of a model field that
 * is a relationship. It is used by applyJoins and applyWhereConditions
 * to build the correct query.
 * @typedef {{
 *  relationshipType: "belongsTo" | "hasMany",
 *  foreignKey: string,
 *  referencesTable: string,
 * }} RelationshipConfig
 *
 * TableConfig is an object where the keys are relationship field names
 * (which don't exist in the database) and the values are RelationshipConfig
 * objects describing that relationship.
 * @typedef {Object.<string, RelationshipConfig} TableConfig
 *
 * TableConfigMap is mapping of database table names to TableConfig objects
 * @typedef {Object.<string, TableConfig>} TableConfigMap
 */

 function setAuditConfig(transaction) {
  let auditStore = auditDataInstance.getStore();
  const identityId = auditStore.identityId;
  console.log(identityId);

   sql`select set_config('audit.identity_id', ${identityId}, true);`.execute(transaction);
}

class ModelAPI {
  /**
   * @param {string} tableName The name of the table this API is for
   * @param {Function} _ Used to be a function that returns the default values for a row in this table. No longer used.
   * @param {TableConfigMap} tableConfigMap
   */
  constructor(tableName, _, tableConfigMap = {}) {
    this._tableName = tableName;
    this._tableConfigMap = tableConfigMap;
    this._modelName = upperCamelCase(this._tableName);
  }

  // async setAuditConfig(transaction) {
 
  //   let auditStore = auditDataInstance.getStore();
  //   console.log(auditStore);
  //   await sql`select set_config('audit.identity_id', '`+auditStore.identityId+`', true);`.execute(transaction);
  // }

  async create(values) {
    const name = tracing.spanNameForModelAPI(this._modelName, "create");
    const db = useDatabase();

    

    return tracing.withSpan(name, async (span) => {
      try { 
        const row = db.transaction().execute(async (transaction) => {

           setAuditConfig(transaction);

          const query = transaction
            .insertInto(this._tableName)
            .values(
              snakeCaseObject({
                ...values,
              })
            )
            .returningAll();

            span.setAttribute("sql", query.compile().sql);
           return await query.executeTakeFirstOrThrow();
        });

        return camelCaseObject(row);
      } catch (e) {
        throw new DatabaseError(e);
      }
    });
  }

  async findOne(where = {}) {
    const name = tracing.spanNameForModelAPI(this._modelName, "findOne");
    const db = useDatabase();

    return tracing.withSpan(name, async (span) => {
      let builder = db
        .selectFrom(this._tableName)
        .distinctOn(`${this._tableName}.id`)
        .selectAll(this._tableName);

      const context = new QueryContext([this._tableName], this._tableConfigMap);

      builder = applyJoins(context, builder, where);
      builder = applyWhereConditions(context, builder, where);

      span.setAttribute("sql", builder.compile().sql);
      const row = await builder.executeTakeFirst();
      if (!row) {
        return null;
      }

      return camelCaseObject(row);
    });
  }

  async findMany(params) {
    const name = tracing.spanNameForModelAPI(this._modelName, "findMany");
    const db = useDatabase();
    const where = params?.where || {};

    return tracing.withSpan(name, async (span) => {
      const context = new QueryContext([this._tableName], this._tableConfigMap);

      let builder = db
        .selectFrom((qb) => {
          // We need to wrap this query as a sub query in the selectFrom because you cannot apply a different order by column when using distinct(id)
          let builder = qb
            .selectFrom(this._tableName)
            .distinctOn(`${this._tableName}.id`)
            .selectAll(this._tableName);

          builder = applyJoins(context, builder, where);
          builder = applyWhereConditions(context, builder, where);

          builder = builder.as(this._tableName);

          return builder;
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

  async update(where, values) {
    const name = tracing.spanNameForModelAPI(this._modelName, "update");
    const db = useDatabase();

    return tracing.withSpan(name, async (span) => {
      let builder = db.updateTable(this._tableName).returningAll();

      builder = builder.set(snakeCaseObject(values));

      const context = new QueryContext([this._tableName], this._tableConfigMap);

      // TODO: support joins for update
      builder = applyWhereConditions(context, builder, where);

      span.setAttribute("sql", builder.compile().sql);

      try {
        const row = await builder.executeTakeFirstOrThrow();
        return camelCaseObject(row);
      } catch (e) {
        throw new DatabaseError(e);
      }
    });
  }

  async delete(where) {
    const name = tracing.spanNameForModelAPI(this._modelName, "delete");
    const db = useDatabase();

    return tracing.withSpan(name, async (span) => {
      let builder = db.deleteFrom(this._tableName).returning(["id"]);

      const context = new QueryContext([this._tableName], this._tableConfigMap);

      // TODO: support joins for delete
      builder = applyWhereConditions(context, builder, where);

      span.setAttribute("sql", builder.compile().sql);
      try {
        const row = await builder.executeTakeFirstOrThrow();
        return row.id;
      } catch (e) {
        throw new DatabaseError(e);
      }
    });
  }

  where(where) {
    const db = useDatabase();

    let builder = db
      .selectFrom(this._tableName)
      .distinctOn(`${this._tableName}.id`)
      .selectAll(this._tableName);

    const context = new QueryContext([this._tableName], this._tableConfigMap);

    builder = applyJoins(context, builder, where);
    builder = applyWhereConditions(context, builder, where);

    return new QueryBuilder(this._tableName, context, builder);
  }
}

module.exports = {
  ModelAPI,
  DatabaseError,
};
