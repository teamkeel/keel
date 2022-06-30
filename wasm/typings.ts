export interface GoExec {
	run: (instance: WebAssembly.Instance) => void
	importObject: any
}

export interface KeelAPI {
  format: (schemaString: string) => string
  validate: (schemaString: string, options?: ValidateOptions) => any
}

export interface ValidateOptions {
  color: boolean
}

export interface Pos {
	column: number
	filename: string
	line: number
	offset: number
}

export interface ValidationError {
  code: string
  pos: Pos
  endPos: Pos
  hint: string
  shortMessage: string
  message: string
}

export interface ValidationResult {
  errors: ValidationError[]
  ast: any
}