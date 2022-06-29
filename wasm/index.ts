import wasm from './keel.wasm'
import "./lib/wasm_exec_node.js"

interface GoExec {
	run: (instance: WebAssembly.Instance) => void
	importObject: any
}

interface ValidateOptions {
  color: boolean
}

interface KeelAPI {
  format: (schemaString: string) => string
  validate: (schemaString: string, options?: ValidateOptions) => string
}

const instantiate = async () : Promise<KeelAPI> => {
  const go: GoExec = new (globalThis as any).Go();
  const { instance } = await WebAssembly.instantiate(wasm, go.importObject);
  go.run(instance);
  const keel: KeelAPI = (globalThis as any).keel;

  return keel;
};

// usage: instantiate().then((keel) => console.log(keel.validate("model   Post {}")))

export default instantiate
