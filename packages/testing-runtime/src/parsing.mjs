import { File } from "@teamkeel/functions-runtime";

export async function parseInputs(inputs) {
  if (inputs != null && typeof inputs === "object") {
    for (const keys of Object.keys(inputs)) {
      if (inputs[keys] !== null && typeof inputs[keys] === "object") {
        if (isDuration(inputs[keys])) {
          inputs[keys] = inputs[keys].toISOString();
        } else if (isInlineFileOrFile(inputs[keys])) {
          const contents = await inputs[keys].read();
          inputs[keys] = `data:${inputs[keys].contentType};name=${
            inputs[keys].filename
          };base64,${contents.toString("base64")}`;
        } else {
          inputs[keys] = await parseInputs(inputs[keys]);
        }
      }
    }
  }
  return inputs;
}

function isInlineFileOrFile(obj) {
  return (
    obj &&
    typeof obj === "object" &&
    (obj.constructor.name === "InlineFile" || obj.constructor.name === "File")
  );
}

function isDuration(obj) {
  return obj && typeof obj === "object" && obj.constructor.name === "Duration";
}

export function parseOutputs(data) {
  if (!data) {
    return null;
  }

  if (!isPlainObject(data)) {
    return data;
  }

  const keys = data ? Object.keys(data) : [];
  const row = {};

  for (const key of keys) {
    const value = data[key];

    if (isPlainObject(value)) {
      if (value.key && value.size && value.filename && value.contentType) {
        row[key] = File.fromDbRecord(value);
      } else {
        row[key] = parseOutputs(value);
      }
    } else if (
      Array.isArray(value) &&
      value.every((item) => typeof item === "object" && item !== null)
    ) {
      const arr = [];
      for (let item of value) {
        if (item.key && item.size && item.filename && item.contentType) {
          arr.push(File.fromDbRecord(item));
        } else {
          arr.push(parseOutputs(item));
        }
      }
      row[key] = arr;
    } else {
      row[key] = value;
    }
  }
  return row;
}

function isPlainObject(obj) {
  return Object.prototype.toString.call(obj) === "[object Object]";
}

const dateFormat =
  /^\d{4}-\d{2}-\d{2}(?:T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:?\d{2})?)?$/;

export function reviver(key, value) {
  // Handle date strings
  if (typeof value === "string") {
    if (dateFormat.test(value)) {
      return new Date(value);
    }
  }

  // Handle nested objects
  if (value !== null && typeof value === "object") {
    // Handle arrays
    if (Array.isArray(value)) {
      return value.map((item) => {
        if (typeof item === "string") {
          if (dateFormat.test(item)) {
            return new Date(item);
          }
        }
        return item;
      });
    }

    // Handle plain objects
    for (const k in value) {
      if (typeof value[k] === "string") {
        if (dateFormat.test(value[k])) {
          value[k] = new Date(value[k]);
        }
      }
    }
  }

  return value;
}
