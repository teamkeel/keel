import { Duration } from "./Duration";

function isPlainObject(obj: unknown): boolean {
  return Object.prototype.toString.call(obj) === "[object Object]";
}

function isRichType(obj: unknown): boolean {
  if (!isPlainObject(obj)) {
    return false;
  }

  return obj instanceof Duration;
}

export { isPlainObject, isRichType };
