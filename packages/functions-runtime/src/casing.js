import { snakeCase, camelCase, splitSeparateNumbers } from "change-case";

function camelCaseObject(obj = {}) {
  const r = {};
  for (const key of Object.keys(obj)) {
    r[
      camelCase(key, {
        split: splitSeparateNumbers,
        mergeAmbiguousCharacters: true,
      })
    ] = obj[key];
  }
  return r;
}

function snakeCaseObject(obj) {
  const r = {};
  for (const key of Object.keys(obj)) {
    r[
      snakeCase(key, {
        split: splitSeparateNumbers,
      })
    ] = obj[key];
  }
  return r;
}

function upperCamelCase(s) {
  s = camelCase(s);
  return s[0].toUpperCase() + s.substring(1);
}

module.exports = {
  camelCaseObject,
  snakeCaseObject,
  snakeCase,
  camelCase,
  upperCamelCase,
};
