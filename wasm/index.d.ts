declare module '@teamkeel/wasm/build' {
  export {};

}
declare module '@teamkeel/wasm/index' {
  import { KeelAPI } from '@teamkeel/wasm/typings';
  const keel: () => Promise<KeelAPI>;
  export default keel;

}
declare module '@teamkeel/wasm/lib/transformKeys' {
  const _default: (obj: Record<string, any>) => Record<string, any>;
  export default _default;

}
declare module '@teamkeel/wasm/lib/wasm_exec' {

}
declare module '@teamkeel/wasm/lib/wasm_exec_node' {
  export {};

}
declare module '@teamkeel/wasm/typings' {
  export interface GoExec {
      run: (instance: WebAssembly.Instance) => void;
      importObject: any;
  }
  export interface KeelAPI {
      format: (schemaString: string) => string;
      validate: (schemaString: string, options?: ValidateOptions) => ValidationResult;
  }
  export interface ValidateOptions {
      color: boolean;
  }
  export interface Pos {
      column: number;
      filename: string;
      line: number;
      offset: number;
  }
  export interface ValidationError {
      code: string;
      pos: Pos;
      endPos: Pos;
      hint: string;
      shortMessage: string;
      message: string;
  }
  export interface ValidationResult {
      errors: ValidationError[];
      ast: any;
  }

}
declare module '@teamkeel/wasm' {
  import main = require('@teamkeel/wasm/index');
  export = main;
}