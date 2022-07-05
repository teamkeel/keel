export interface GoExec {
	run: (instance: WebAssembly.Instance) => void
	importObject: any
}

export interface KeelAPI {
  format: (schemaString: string) => string
  validate: (schemaString: string, options: ValidateOptions) => ValidationResult
  completions: (schemaString: string, position: SimplePosition) => CompletionResult
}

export interface ValidateOptions {
  color: boolean
}

export interface SimplePosition {
	column: number
	line: number
}

export interface Position extends SimplePosition {
	filename: string
	offset: number
}

export interface CompletionItem {
  description: string
  label: string
  node: Node
}

export interface CompletionResult {
  completions: CompletionItem[]
  ast: any
}

export interface ValidationError {
  code: string
  pos: Position
  endPos: Position
  hint: string
  shortMessage: string
  message: string
}

export interface ValidationResult {
  errors: ValidationError[]
  ast: any
}