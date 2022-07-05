import wasm from './keel.wasm'
import { GoExec, KeelAPI, ValidationResult, ValidateOptions, ValidationError, CompletionResult, SimplePosition } from './typings'
import transformKeys from './lib/transformKeys';
import util from 'util'

const instantiate = async () : Promise<KeelAPI> => {
  // necessary to do dynamic import of js file here to avoid relative import error
  // in ambient module: https://github.com/piotrwitek/react-redux-typescript-guide/issues/137#issuecomment-805109873
  await import("./lib/wasm_exec_node.js")
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

  const completions = (schemaString: string, position: SimplePosition) : CompletionResult => {
    return api.completions(schemaString, position);
  }

  return { validate, format, completions }
}

// keel().then((api) => {
//   const { completions } = api.completions(
//     `model Post {
//       fields {

//       }
//     }`,
//     { line: 1, column: 1 }
//   );

//   console.log(util.inspect(completions, {showHidden: false, depth: null, colors: true}))
// })

export default keel
