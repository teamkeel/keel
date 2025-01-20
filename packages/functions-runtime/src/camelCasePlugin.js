const { CamelCasePlugin } = require("kysely");
const { isPlainObject, isRichType } = require("./parsing");
// KeelCamelCasePlugin is a wrapper around kysely's camel case plugin.
class KeelCamelCasePlugin {
  constructor(opt) {
    this.opt = opt;
    this.CamelCasePlugin = new CamelCasePlugin(opt);
  }

  transformQuery(args) {
    return this.CamelCasePlugin.transformQuery(args);
  }

  async transformResult(args) {
    if (args.result.rows && Array.isArray(args.result.rows)) {
      return {
        ...args.result,
        rows: args.result.rows.map((row) => this.mapRow(row)),
      };
    }
    return args.result;
  }
  mapRow(row) {
    return Object.keys(row).reduce((obj, key) => {
      let value = row[key];
      if (Array.isArray(value)) {
        value = value.map((it) =>
          canMap(it, this.opt) ? this.mapRow(it) : it
        );
      } else if (canMap(value, this.opt)) {
        value = this.mapRow(value);
      }
      obj[this.CamelCasePlugin.camelCase(key)] = value;
      return obj;
    }, {});
  }
}

function canMap(obj, opt) {
  return (
    isPlainObject(obj) && !opt?.maintainNestedObjectKeys && !isRichType(obj)
  );
}

module.exports.KeelCamelCasePlugin = KeelCamelCasePlugin;
