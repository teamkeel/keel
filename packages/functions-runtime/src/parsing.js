const { InlineFile } = require("./InlineFile");

// parseParams takes a set of inputs and creates objects for the ones that are of a complex type.
//
// inputs that are objects and contain a "__typename" field are resolved to instances of the complex type
// they represent. At the moment, the only supported type is `InlineFile`
function parseParams(inputs) {
  if (inputs != null) {
    Object.keys(inputs).forEach((i) => {
      if (inputs[i] !== null && typeof inputs[i] === "object") {
        if ("__typename" in inputs[i]) {
          switch (inputs[i].__typename) {
            case "InlineFile":
              inputs[i] = InlineFile.fromObject(inputs[i]);
              break;

            default:
              break;
          }
        }
      }
    });
  }

  return inputs;
}

module.exports.parseParams = parseParams;
