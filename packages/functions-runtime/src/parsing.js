const { InlineFile, StoredFile } = require("./InlineFile");

// parseParams takes a set of inputs and creates objects for the ones that are of a complex type.
//
// inputs that are objects and contain a "__typename" field are resolved to instances of the complex type
// they represent. At the moment, the only supported type is `InlineFile`
function parseParams(inputs) {
  if (inputs != null && typeof inputs === "object") {
    Object.keys(inputs).forEach((i) => {
      if (inputs[i] !== null && typeof inputs[i] === "object") {
        if ("__typename" in inputs[i]) {
          switch (inputs[i].__typename) {
            case "InlineFile"://TODO: Stored file???
              inputs[i] = InlineFile.fromDataURL(inputs[i].dataURL);
              break;

            default:
              break;
          }
        } else {
          inputs[i] = parseParams(inputs[i]);
        }
      }
    });
  }

  return inputs;
}

// Iterate through the given object's keys and if any of the values are a rich data type, instantiate their respective class
function transformRichDataTypes(data) {
  const keys = data ? Object.keys(data) : [];
  const row = {};

  for (const key of keys) {
    const value = data[key];
    if (isPlainObject(value)) {
      // if we've got a StoredFile...
      if (value.key && value.size && value.filename && value.contentType) {
        row[key] = StoredFile.fromDbRecord(value);
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

function isReferencingExistingRecord(value) {
  return Object.keys(value).length === 1 && value.id;
}

module.exports = {
  parseParams,
  transformRichDataTypes,
  isPlainObject,
  isReferencingExistingRecord,
};
