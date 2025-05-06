/**
 * QueryContext is used to store state about the current query, for example
 * which joins have already been applied. It is used by applyJoins and
 * applyWhereConditions to generate consistent table aliases for joins.
 *
 * This class has the concept of a "table path". This is just a list of tables, starting
 * with some "root" table and ending with the table we're currently joining to. So
 * for example if we started with a "product" table and joined from there to "order_item"
 * and then to "order" and then to "customer" the table path would be:
 *   ["product", "order_item", "order", "customer"]
 * At this point the "current" table is "customer" and it's alias would be:
 *   "product$order_item$order$customer"
 */
class QueryContext {
  /**
   * @param {string[]} tablePath This is the path from the "root" table to the "current table".
   * @param {import("./ModelAPI").TableConfigMap} tableConfigMap
   * @param {string[]} joins
   */
  constructor(tablePath, tableConfigMap, joins = []) {
    this._tablePath = tablePath;
    this._tableConfigMap = tableConfigMap;
    this._joins = joins;
  }

  clone() {
    return new QueryContext([...this._tablePath], this._tableConfigMap, [
      ...this._joins,
    ]);
  }

  /**
   * Returns true if, given the current table path, a join to the given
   * table has already been added.
   * @param {string} table
   * @returns {boolean}
   */
  hasJoin(table) {
    const alias = joinAlias([...this._tablePath, table]);
    return this._joins.includes(alias);
  }

  /**
   * Adds table to the QueryContext's path and registers the join,
   * calls fn, then pops the table off the path.
   * @param {string} table
   * @param {Function} fn
   */
  withJoin(table, fn) {
    this._tablePath.push(table);
    this._joins.push(this.tableAlias());

    fn();

    // Don't change the _joins list, we want to remember those
    this._tablePath.pop();
  }

  /**
   * Returns the alias that will be used for the current table
   * @returns {string}
   */
  tableAlias() {
    return joinAlias(this._tablePath);
  }

  /**
   * Returns the current table name
   * @returns {string}
   */
  tableName() {
    return this._tablePath[this._tablePath.length - 1];
  }

  /**
   * Return the TableConfig for the current table
   * @returns {import("./ModelAPI").TableConfig | undefined}
   */
  tableConfig() {
    return this._tableConfigMap[this.tableName()];
  }
}

function joinAlias(tablePath) {
  return tablePath.join("$");
}

export { QueryContext };
