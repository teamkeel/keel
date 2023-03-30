const { getDatabase } = require("./database");
const { QueryBuilder } = require("./QueryBuilder");
const { QueryContext } = require("./QueryContext");
const { applyWhereConditions } = require("./applyWhereConditions");
const { applyJoins } = require("./applyJoins");
const { camelCaseObject, snakeCaseObject } = require("./casing");

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

class DatabaseError extends Error {
  constructor(error) {
    super(error.message);
    this.error = error;
  }
}

class ModelAPI {
  /**
   * @param {string} tableName The name of the table this API is for
   * @param {Function} defaultValues A function that returns the default values for a row in this table
   * @param {import("kysely").Kysely} db
   * @param {TableConfigMap} tableConfigMap
   */
  constructor(tableName, defaultValues, db, tableConfigMap = {}) {
    this._db = db || getDatabase();
    this._defaultValues = defaultValues;
    this._tableName = tableName;
    this._tableConfigMap = tableConfigMap;
  }

  async create(values) {
    try {
      const defaults = this._defaultValues();
      const row = await this._db
        .insertInto(this._tableName)
        .values(
          snakeCaseObject({
            ...defaults,
            ...values,
          })
        )
        .returningAll()
        .executeTakeFirstOrThrow();

      return camelCaseObject(row);
    } catch (e) {
      throw new DatabaseError(e);
    }
  }

  async findOne(where = {}) {
    let builder = this._db
      .selectFrom(this._tableName)
      .distinctOn(`${this._tableName}.id`)
      .selectAll(this._tableName);

    const context = new QueryContext([this._tableName], this._tableConfigMap);

    builder = applyJoins(context, builder, where);
    builder = applyWhereConditions(context, builder, where);

    const row = await builder.executeTakeFirst();
    if (!row) {
      return null;
    }

    return camelCaseObject(row);
  }

  async findMany(where = {}) {
    let builder = this._db
      .selectFrom(this._tableName)
      .distinctOn(`${this._tableName}.id`)
      .selectAll(this._tableName);

    const context = new QueryContext([this._tableName], this._tableConfigMap);

    builder = applyJoins(context, builder, where);
    builder = applyWhereConditions(context, builder, where);

    const rows = await builder.orderBy("id").execute();
    return rows.map((x) => camelCaseObject(x));
  }

  async update(where, values) {
    let builder = this._db.updateTable(this._tableName).returningAll();

    builder = builder.set(snakeCaseObject(values));

    const context = new QueryContext([this._tableName], this._tableConfigMap);

    // TODO: support joins for update
    builder = applyWhereConditions(context, builder, where);

    try {
      const row = await builder.executeTakeFirstOrThrow();

      return camelCaseObject(row);
    } catch (e) {
      throw new DatabaseError(e);
    }
  }

  async delete(where) {
    let builder = this._db.deleteFrom(this._tableName).returning(["id"]);

    const context = new QueryContext([this._tableName], this._tableConfigMap);

    // TODO: support joins for delete
    builder = applyWhereConditions(context, builder, where);

    try {
      const row = await builder.executeTakeFirstOrThrow();

      return row.id;
    } catch (e) {
      throw new DatabaseError(e);
    }
  }

  where(where) {
    let builder = this._db
      .selectFrom(this._tableName)
      .distinctOn(`${this._tableName}.id`)
      .selectAll(this._tableName);

    const context = new QueryContext([this._tableName], this._tableConfigMap);

    builder = applyJoins(context, builder, where);
    builder = applyWhereConditions(context, builder, where);

    return new QueryBuilder(context, builder);
  }
}

module.exports = {
  ModelAPI,
  DatabaseError,
};
