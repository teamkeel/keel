import { snakeCase, camelCase } from "change-case";

function camelCaseObject(obj = {}) {
  const r = {};
  for (const key of Object.keys(obj)) {
    r[
      camelCase(key, {
        transform: camelCaseTransform,
        splitRegexp: [
          /([a-z0-9])([A-Z])/g,
          /([A-Z])([A-Z][a-z])/g,
          /([a-zA-Z])([0-9])/g,
        ],
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
        splitRegexp: [
          /([a-z0-9])([A-Z])/g,
          /([A-Z])([A-Z][a-z])/g,
          /([a-zA-Z])([0-9])/g,
        ],
      })
    ] = obj[key];
  }
  return r;
}

function upperCamelCase(s) {
  s = camelCase(s);
  return s[0].toUpperCase() + s.substring(1);
}

function camelCaseTransform(input, index) {
  if (index === 0) return input.toLowerCase();
  const firstChar = input.charAt(0);
  const lowerChars = input.substr(1).toLowerCase();
  return `${firstChar.toUpperCase()}${lowerChars}`;
}

export {
  camelCaseObject,
  snakeCaseObject,
  snakeCase,
  camelCase,
  upperCamelCase,
};
