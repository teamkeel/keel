const { snakeCase, camelCase } = require("change-case");

function camelCaseObject(obj) {
  const r = {};
  for (const key of Object.keys(obj)) {
    r[camelCase(key)] = obj[key];
  }
  return r;
}

function snakeCaseObject(obj) {
  const r = {};
  for (const key of Object.keys(obj)) {
    r[snakeCase(key)] = obj[key];
  }
  return r;
}

module.exports = {
  camelCaseObject,
  snakeCaseObject,
  snakeCase,
  camelCase,
};
