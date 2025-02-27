const { Duration } = require("./Duration");

function isPlainObject(obj) {
  return Object.prototype.toString.call(obj) === "[object Object]";
}

function isRichType(obj) {
  if (!isPlainObject(obj)) {
    return false;
  }

  return obj instanceof Duration;
}

module.exports = {
  isPlainObject,
  isRichType,
};
