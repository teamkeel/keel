/**
 * The available hooks for a 'get' function
 * @typeParam M - The Model this function is declared in
 * @typeParam QB - The QueryBuilder type for the model
 * @typeParam I - The function inputs
 */
type GetFunctionHooks<M, QB, I> = {
  beforeQuery?: (
    ctx: ContextAPI,
    inputs: I,
    query: QB
  ) => Promise<QB | M | null | Error> | QB | M | null | Error;
  afterQuery?: (
    ctx: ContextAPI,
    inputs: I,
    record: M
  ) => Promise<M | Error> | M | Error;
};

/**
 * The available hooks for a 'list' function
 * @typeParam M - The Model this function is declared in
 * @typeParam QB - The QueryBuilder type for the model
 * @typeParam I - The function inputs
 */
type ListFunctionHooks<M, QB, I> = {
  beforeQuery?: (
    ctx: ContextAPI,
    inputs: I,
    query: QB
  ) => Promise<QB | Array<M> | Error> | QB | Array<M> | Error;
  afterQuery?: (
    ctx: ContextAPI,
    inputs: I,
    records: Array<M>
  ) => Promise<Array<M> | Error> | Array<M> | Error;
};

/**
 * The available hooks for a 'create' function
 * @typeParam M - The Model this function is declared in
 * @typeParam QB - The QueryBuilder type for the model
 * @typeParam I - The function inputs
 * @typeParam I - The values that have been derived from the inputs
 * @typeParam C - The values that will be used to create an M record
 */
type CreateFunctionHooks<M, QB, I, V, C> = {
  beforeWrite?: (
    ctx: ContextAPI,
    inputs: I,
    values: V
  ) => Promise<C | Error> | C | Error;
  afterWrite?: (
    ctx: ContextAPI,
    inputs: I,
    data: M
  ) => Promise<M | void | Error> | M | void | Error;
};

/**
 * The available hooks for a 'create' function
 * @typeParam M - The Model this function is declared in
 * @typeParam QB - The QueryBuilder type for the model
 * @typeParam I - The function inputs
 * @typeParam C - The values that can be used to update an M record
 */
type UpdateFunctionHooks<M, QB, I, V> = {
  beforeQuery?: (
    ctx: ContextAPI,
    inputs: I,
    query: QB
  ) => Promise<M | QB | Error> | M | QB | Error;
  beforeWrite?: (
    ctx: ContextAPI,
    inputs: I,
    values: V,
    record: M
  ) => Promise<Partial<M> | Error> | Partial<M> | Error;
  afterWrite?: (
    ctx: ContextAPI,
    inputs: I,
    data: M
  ) => Promise<M | void | Error> | M | void | Error;
};

/**
 * The available hooks for a 'delete' function
 * @typeParam M - The Model this function is declared in
 * @typeParam QB - The QueryBuilder type for the model
 * @typeParam I - The function inputs
 */
type DeleteFunctionHooks<M, QB, I> = {
  beforeQuery?: (
    ctx: ContextAPI,
    inputs: I,
    query: QB
  ) => Promise<M | QB | Error> | M | QB | Error;
  beforeWrite?: (
    ctx: ContextAPI,
    inputs: I,
    record: M
  ) => Promise<void | Error> | void | Error;
  afterWrite?: (
    ctx: ContextAPI,
    inputs: I,
    data: M
  ) => Promise<void | Error> | void | Error;
};
