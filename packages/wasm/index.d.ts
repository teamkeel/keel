export function format(schema: string): Promise<string>;

export function validate(req: ValidateRequest): Promise<ValidationResult>;

export function completions(
  req: GetCompletionsRequest
): Promise<CompletionResult>;

export function getDefinition(
  req: GetDefinitionRequest
): Promise<DefinitionResult>;

export interface DefinitionResult {
  schema?: Position;
  function?: { name: string };
}

export interface SchemaDefinition {
  schema: SchemaDefinition;
}

export interface GetCompletionsRequest {
  position: Position;
  schemaFiles: SchemaFile[];
  config?: string;
}

export interface GetDefinitionRequest {
  position: Position;
  schemaFiles: SchemaFile[];
}

export interface ValidateRequest {
  schemaFiles: SchemaFile[];
  config?: string;
  includeWarnings?: bool;
}

export interface SchemaFile {
  filename: string;
  contents: string;
}

export interface Position {
  filename: string;
  line: number;
  column: number;
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
  message: string;
}

export interface ValidationResult {
  errors: ValidationError[];
  warnings?: ValidationError[];
}
