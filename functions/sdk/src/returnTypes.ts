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

export interface FunctionCreateResponse<T> {
  object?: T;
  errors?: FunctionError[];
}

export interface FunctionGetResponse<T> {
  object?: T;

  // todo: it doesnt make sense for ValidationError to be in the union below
  errors?: FunctionError[];
}

export interface FunctionDeleteResponse<T> {
  success: boolean;
  errors?: FunctionError[];
}

export interface FunctionListResponse<T> {
  collection: T[];
  errors?: FunctionError[];
  // todo: add type for pagination
}

export interface FunctionUpdateResponse<T> {
  object?: T;
  errors?: FunctionError[];
}

export interface FunctionAuthenticateResponse<T> {
  identityId?: string;
  identityCreated: boolean;
  errors?: FunctionError[];
}
