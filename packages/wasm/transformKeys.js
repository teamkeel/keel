const deepMapKeys = require("deep-map-keys");

function snakeToCamel(key) {
  return key.replace(/_(\w)/g, (match, char) => char.toUpperCase());
}

module.exports.transformKeys = (obj) => {
  return deepMapKeys(obj, (key, value) => snakeToCamel(key));
};
