import wasm from './keel.wasm'
import { GoExec, KeelAPI, ValidationResult, ValidateOptions, ValidationError, CompletionResult, SimplePosition } from './typings'
import transformKeys from './lib/transformKeys';

const instantiate = async () : Promise<KeelAPI> => {
  // necessary to do dynamic import of js file here to avoid relative import error
  // in ambient module: https://github.com/piotrwitek/react-redux-typescript-guide/issues/137#issuecomment-805109873
  await import("./lib/wasm_exec_node.js")
  const go: GoExec = new (globalThis as any).Go();
  const { instance } = await WebAssembly.instantiate(wasm, go.importObject);
  go.run(instance);

  return (globalThis as any).keel as KeelAPI;
};

const keel = () : KeelAPI => {
  const validate = async (schemaString: string, opts: ValidateOptions) : Promise<ValidationResult> => {
    const api = await instantiate();

    const result = api.validate(schemaString, opts) as any;

    if (result.error) {
      return {
        errors: [result.error],
        ast: null
      }
    }

    if (!result || !result.validationErrors) {
      return {
        errors: [],
        ast: null
      }
    }

    const { validationErrors: { Errors: errors }, ast } = result

    const transformedErrors = (errors || []).map((err: ValidationError) => transformKeys(err));

    return {
      errors: transformedErrors,
      ast: ast
    }
  }

  const format = async (schemaString: string) : Promise<string> => {
    const api = await instantiate();

    return api.format(schemaString);
  }

  const completions = async (schemaString: string, position: SimplePosition) : Promise<CompletionResult> => {
    const api = await instantiate();

    return api.completions(schemaString, position);
  }

  return { validate, format, completions }
}

export default keel
