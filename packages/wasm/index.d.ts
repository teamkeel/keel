export function format(schema: string): Promise<string>;

export function validate(schemaString: string): Promise<ValidationResult>;

export function completions(
  schemaString: string,
  position: SimplePosition
): Promise<CompletionResult>;

export interface SimplePosition {
  column: number;
  line: number;
}

export interface Position extends SimplePosition {
  filename: string;
  offset: number;
}

export interface CompletionItem {
  description: string;
  label: string;
  insertText: string;
  kind: string;
}

export interface CompletionResult {
  completions: CompletionItem[];
}

export interface ValidationError {
  code: string;
  pos: Position;
  endPos: Position;
  hint: string;
  shortMessage: string;
  message: string;
}

export interface ValidationResult {
  errors: ValidationError[];
}
