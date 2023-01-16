const { applyWhereConditions } = require("./applyWhereConditions");
const { camelCaseObject, snakeCaseObject } = require("./casing");
const { QueryBuilder } = require("./QueryBuilder");
const { getDatabase } = require("./database");

class ModelAPI {
  constructor(tableName, defaultValues, db) {
    this._tableName = tableName;
    this._defaultValues = defaultValues;
    this._db = db || getDatabase();
  }

  async create(values) {
    const row = await this._db
      .insertInto(this._tableName)
      .values(
        snakeCaseObject({
          ...this._defaultValues(),
          ...values,
        })
      )
      .returningAll()
      .executeTakeFirst();
    return camelCaseObject(row);
  }

  async findOne(where) {
    const row = await this._db
      .selectFrom(this._tableName)
      .selectAll()
      .where((qb) => {
        return applyWhereConditions(qb, where);
      })
      .executeTakeFirst();
    if (!row) {
      return null;
    }
    return camelCaseObject(row);
  }

  async findMany(where) {
    const rows = await this._db
      .selectFrom(this._tableName)
      .selectAll()
      .where((qb) => {
        return applyWhereConditions(qb, where);
      })
      .execute();
    return rows.map((x) => camelCaseObject(x));
  }

  async update(where, values) {
    const row = await this._db
      .updateTable(this._tableName)
      .returningAll()
      .set(snakeCaseObject(values))
      .where((qb) => {
        return applyWhereConditions(qb, where);
      })
      .executeTakeFirstOrThrow();
    return camelCaseObject(row);
  }

  async delete(where) {
    const row = await this._db
      .deleteFrom(this._tableName)
      .returning(["id"])
      .where((qb) => {
        return applyWhereConditions(qb, where);
      })
      .executeTakeFirstOrThrow();
    return row.id;
  }

  where(conditions) {
    const q = this._db
      .selectFrom(this._tableName)
      .selectAll()
      .where((qb) => {
        return applyWhereConditions(qb, conditions);
      });
    return new QueryBuilder(q);
  }
}

module.exports.ModelAPI = ModelAPI;
