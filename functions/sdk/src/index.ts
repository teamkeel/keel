import Query, { ChainableQuery } from "./query";
import * as QueryConstraints from "./constraints";
import Logger, { ConsoleTransport, Level as LogLevel } from "./logger";
import { Identity } from "./types";

// ../../client doesn't exist until the runtime codegens it when the run
// command is executed.
// It will exist in the node_modules dir at @teamkeel/client when
// everything has been code generated.
// For the meantime, we want to ts-ignore this export as it will cause esbuild
// to fail
//@ts-ignore
export * from "../../client";

// export all of the generic return types that are used by custom functions
// these include different types of responses for different operations
export * from "./returnTypes";

export * from "./db/resolver";

import { queryResolverFromEnv } from "./db/resolver";

export {
  Query,
  ChainableQuery,
  QueryConstraints,
  Logger,
  ConsoleTransport,
  LogLevel,
  Identity,
  queryResolverFromEnv,
};
