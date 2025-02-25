const { Duration } = require("./Duration");
const { InlineFile, File } = require("./File");
const { isPlainObject } = require("./type-utils");

// parseInputs takes a set of inputs and creates objects for the ones that are of a complex type.
//
// inputs that are objects and contain a "__typename" field are resolved to instances of the complex type
// they represent.
function parseInputs(inputs) {
  if (inputs != null && typeof inputs === "object") {
    for (const k of Object.keys(inputs)) {
      if (inputs[k] !== null && typeof inputs[k] === "object") {
        if (Array.isArray(inputs[k])) {
          inputs[k] = inputs[k].map((item) => {
            if (item && typeof item === "object" && "__typename" in item) {
              switch (item.__typename) {
                case "InlineFile":
                  return InlineFile.fromDataURL(item.dataURL);
                case "Duration":
                  return Duration.fromISOString(item.interval);
                default:
                  return item;
              }
            }
            return item;
          });
        } else if ("__typename" in inputs[k]) {
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
async function parseOutputs(outputs) {
  if (outputs != null && typeof outputs === "object") {
    for (const k of Object.keys(outputs)) {
      if (outputs[k] !== null && typeof outputs[k] === "object") {
        if (Array.isArray(outputs[k])) {
          outputs[k] = await Promise.all(
            outputs[k].map((item) => parseOutputs(item))
          );
        } else if (outputs[k] instanceof InlineFile) {
          const stored = await outputs[k].store();
          outputs[k] = stored;
        } else if (outputs[k] instanceof Duration) {
          outputs[k] = outputs[k].toISOString();
        } else {
          outputs[k] = await parseOutputs(outputs[k]);
        }
      }
    }
  }

  return outputs;
}

// transformRichDataTypes iterates through the given object's keys and if any of the values are a rich data type, instantiate their respective class
function transformRichDataTypes(data) {
  const keys = data ? Object.keys(data) : [];
  const row = {};

  for (const key of keys) {
    const value = data[key];
    if (Array.isArray(value)) {
      row[key] = value.map((item) => transformRichDataTypes({ item }).item);
    } else if (isPlainObject(value)) {
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
    } else {
      row[key] = value;
    }
  }

  return row;
}

function isReferencingExistingRecord(value) {
  return Object.keys(value).length === 1 && value.id;
}

module.exports = {
  parseInputs,
  parseOutputs,
  transformRichDataTypes,
  isReferencingExistingRecord,
};
