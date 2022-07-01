import wasm from './keel.wasm'
import { GoExec, KeelAPI, ValidationResult, ValidateOptions, ValidationError } from './typings'
import transformKeys from './lib/transformKeys';

import "./lib/wasm_exec_node.js"

const instantiate = async () : Promise<KeelAPI> => {
  const go: GoExec = new (globalThis as any).Go();
  const { instance } = await WebAssembly.instantiate(wasm, go.importObject);
  go.run(instance);

  return (globalThis as any).keel as KeelAPI;
};

const keel = async () : Promise<KeelAPI> => {
  const api = await instantiate();

  const validate = (schemaString: string, opts: ValidateOptions) : ValidationResult => {
    const result = api.validate(schemaString, opts) as any;
    const { validationErrors: { Errors: errors }, ast } = result

    const transformedErrors = (errors || []).map((err: ValidationError) => transformKeys(err));

    return {
      errors: transformedErrors,
      ast: ast
    }
  }

  const format = (schemaString: string) : string => {
    return api.format(schemaString);
  }

  return { validate, format }
}

export default keel
