// Needed for wasm_exec
globalThis.crypto = require("crypto");

const { transformKeys } = require("./transformKeys");
const { wasm } = require("./dist/wasm.js");
require("./lib/wasm_exec");

const instantiate = async () => {
  const go = new globalThis.Go();

  const wasmModule = await WebAssembly.instantiate(wasm, go.importObject);
  go.run(wasmModule.instance);

  return globalThis.keel;
};

const keel = () => {
  const validate = async (schemaString, opts) => {
    const api = await instantiate();

    const result = api.validate(schemaString, opts || {});

    if (result.error) {
      return {
        errors: [result.error],
        ast: null,
      };
    }

    if (!result || !result.validationErrors) {
      return {
        errors: [result.error],
        ast: null,
      };
    }

    const {
      validationErrors: { Errors: errors },
      ast,
    } = result;

    const transformedErrors = (errors || []).map((err) => transformKeys(err));

    return {
      errors: transformedErrors,
      ast: ast,
    };
  };

  const format = async (schemaString) => {
    const api = await instantiate();

    return api.format(schemaString);
  };

  const completions = async (schemaString, position) => {
    const api = await instantiate();

    return api.completions(schemaString, position);
  };

  return { validate, format, completions };
};

module.exports.keel = keel;
