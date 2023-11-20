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
  ) => Promise<QB | M | null> | QB | M | null;
  afterQuery?: (ctx: ContextAPI, inputs: I, record: M) => Promise<M> | M;
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
  ) => Promise<QB | Array<M>> | QB | Array<M>;
  afterQuery?: (
    ctx: ContextAPI,
    inputs: I,
    records: Array<M>
  ) => Promise<Array<M>> | Array<M>;
};

/**
 * The available hooks for a 'create' function
 * @typeParam M - The Model this function is declared in
 * @typeParam QB - The QueryBuilder type for the model
 * @typeParam I - The function inputs
 * @typeParam C - The values that can be used to create an M record
 */
type CreateFunctionHooks<M, QB, I, C> = {
  beforeWrite?: (ctx: ContextAPI, inputs: I, values: C) => Promise<C> | C;
  afterWrite?: (
    ctx: ContextAPI,
    inputs: I,
    data: M
  ) => Promise<M | void> | M | void;
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
  ) => Promise<M | QB> | M | QB;
  beforeWrite?: (
    ctx: ContextAPI,
    inputs: I,
    values: V,
    record: M
  ) => Promise<Partial<M>> | Partial<M>;
  afterWrite?: (
    ctx: ContextAPI,
    inputs: I,
    data: M
  ) => Promise<M | void> | M | void;
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
  ) => Promise<M | QB> | M | QB;
  beforeWrite?: (ctx: ContextAPI, inputs: I, record: M) => Promise<void> | void;
  afterWrite?: (ctx: ContextAPI, inputs: I, data: M) => Promise<void> | void;
};
