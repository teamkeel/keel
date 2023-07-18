import {
  QueryKey,
  UseInfiniteQueryOptions,
  UseMutationOptions,
  UseQueryOptions,
  useInfiniteQuery,
  useMutation,
  useQuery,
} from "@tanstack/react-query";

type FunctionTypes = "queries" | "mutations";

export const keelQuery = <T extends (...args: any) => any>(useKeel: T) => {
  type KeelType = ReturnType<typeof useKeel>;

  type QueryKeys<F extends FunctionTypes> = keyof KeelType[F];
  type QueryArgs<F extends FunctionTypes, K extends QueryKeys<F>> = Parameters<
    KeelType[F][K]
  >[0];
  type QueryResult<
    F extends FunctionTypes,
    K extends QueryKeys<F>
  > = ReturnType<KeelType[F][K]>;

  type Result<F extends FunctionTypes, K extends QueryKeys<F>> = Exclude<
    Awaited<QueryResult<F, K>>["data"],
    undefined
  >;
  type Error<F extends FunctionTypes, K extends QueryKeys<F>> = Exclude<
    Awaited<QueryResult<F, K>>["error"],
    undefined
  >;

  return {
    useKeelQuery: <F extends "queries", K extends QueryKeys<"queries">>(
      key: K,
      args: QueryArgs<F, K>,
      options?: Omit<UseQueryOptions<Result<F, K>, Error<F, K>>, "queryFn">
    ) => {
      const keel = useKeel();
      return useQuery<Result<F, K>, Error<F, K>>(
        queryKeys(key, args),
        async () => {
          const res = await keel[key](args);
          if (res.error) {
            return Promise.reject(res.error);
          }
          return res.data;
        },
        options
      );
    },
    useInfiniteKeelQuery: <F extends "queries", K extends QueryKeys<"queries">>(
      key: K,
      args: QueryArgs<F, K>,
      options?: Omit<
        UseInfiniteQueryOptions<Result<F, K>, Error<F, K>>,
        "queryFn"
      >
    ) => {
      const keel = useKeel();
      return useInfiniteQuery<Result<F, K>, Error<F, K>>(
        queryKeys(key, args),
        async () => {
          const res = await keel[key](args);
          if (res.error) {
            return Promise.reject(res.error);
          }
          return res.data;
        },
        options
      );
    },
    useKeelMutation: <F extends "mutations", K extends QueryKeys<"mutations">>(
      key: K,
      options?: Omit<
        UseMutationOptions<Result<F, K>, Error<F, K>, QueryArgs<F, K>>,
        "queryFn"
      >
    ) => {
      const keel = useKeel();
      return useMutation<Result<F, K>, Error<F, K>, QueryArgs<F, K>>(
        [key],
        async (args) => {
          const res = await keel[key](args);
          if (res.error) {
            return Promise.reject(res.error);
          }
          return res.data;
        },
        options
      );
    },
  };
};

const queryKeys = (key: any, args?: any): QueryKey => {
  // Query key is either ["action name", {args}] or ["action name", "id", {args}]
  const queryKey = [key];
  if (args && args.id) {
    queryKey.push(args.id);
  }
  if (args) {
    queryKey.push(args);
  }

  return queryKey;
};
