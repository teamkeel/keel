// ValidationErrors will be returned when interacting with
// the Query API (creating, updating entities)
export interface ValidationError {
  field: string;
  message: string;
  code: string;
}

// ExecutionError represents other misc errors
// that can occur during the execution of a custom function
export interface ExecutionError {
  message: string;

  // todo: implement stacks
  stack: string;
}

export type FunctionError = ValidationError | ExecutionError;

export interface FunctionResponse {
  errors?: FunctionError[];
}

export interface FunctionCreateResponse<T> extends FunctionResponse {
  object?: T;
}

export interface FunctionGetResponse<T> extends FunctionResponse {
  object?: T;
}

export interface FunctionDeleteResponse<T> extends FunctionResponse {
  success: boolean;
  errors?: FunctionError[];
}

export interface FunctionListResponse<T> extends FunctionResponse {
  collection: T[];
}

export interface FunctionUpdateResponse<T> extends FunctionResponse {
  object?: T;
}

export interface FunctionAuthenticateResponse {
  identityId?: string;
  identityCreated: boolean;
}
