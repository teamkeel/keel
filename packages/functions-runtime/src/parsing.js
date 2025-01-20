const { Duration } = require("./Duration");
const { InlineFile, File } = require("./File");

// parseInputs takes a set of inputs and creates objects for the ones that are of a complex type.
//
// inputs that are objects and contain a "__typename" field are resolved to instances of the complex type
// they represent. At the moment, the only supported type is `InlineFile`
function parseInputs(inputs) {
  if (inputs != null && typeof inputs === "object") {
    for (const k of Object.keys(inputs)) {
      if (inputs[k] !== null && typeof inputs[k] === "object") {
        if ("__typename" in inputs[k]) {
          switch (inputs[k].__typename) {
            case "InlineFile":
              inputs[k] = InlineFile.fromDataURL(inputs[k].dataURL);
              break;
            case "Duration":
              inputs[k] = Duration.fromISOString(inputs[k].interval);
              break;
            default:
              break;
          }
        } else {
          inputs[k] = parseInputs(inputs[k]);
        }
      }
    }
  }

  return inputs;
}

// parseOutputs will take a response from the custom function and perform operations on any fields if necessary.
//
// For example, InlineFiles need to be stored before returning the response.
async function parseOutputs(inputs) {
  if (inputs != null && typeof inputs === "object") {
    for (const k of Object.keys(inputs)) {
      if (inputs[k] !== null && typeof inputs[k] === "object") {
        if (inputs[k] instanceof InlineFile) {
          const stored = await inputs[k].store();
          inputs[k] = stored;
        } else if (inputs[k] instanceof Duration) {
          inputs[k] = inputs[k].toISOString();
        } else {
          inputs[k] = await parseOutputs(inputs[k]);
        }
      }
    }
  }

  return inputs;
}

// transformRichDataTypes iterates through the given object's keys and if any of the values are a rich data type, instantiate their respective class
function transformRichDataTypes(data) {
  const keys = data ? Object.keys(data) : [];
  const row = {};

  for (const key of keys) {
    const value = data[key];
    if (isPlainObject(value)) {
      if (value._typename == "Duration" && value.pgInterval) {
        row[key] = new Duration(value.pgInterval);
      } else if (
        value.key &&
        value.size &&
        value.filename &&
        value.contentType
      ) {
        row[key] = File.fromDbRecord(value);
      } else {
        row[key] = value;
      }
      continue;
    }

    row[key] = value;
  }

  return row;
}

function isPlainObject(obj) {
  return Object.prototype.toString.call(obj) === "[object Object]";
}

function isRichType(obj) {
  if (!isPlainObject(obj)) {
    return false;
  }

  return obj instanceof Duration;
}

function isReferencingExistingRecord(value) {
  return Object.keys(value).length === 1 && value.id;
}

module.exports = {
  parseInputs,
  parseOutputs,
  transformRichDataTypes,
  isPlainObject,
  isRichType,
  isReferencingExistingRecord,
};
