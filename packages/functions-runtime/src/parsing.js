const { InlineFile } = require("./InlineFile");

function parseParams(inputs) {
  if (inputs != null) {
    Object.keys(inputs).forEach((i) => {
      if (typeof inputs[i] === "object") {
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
