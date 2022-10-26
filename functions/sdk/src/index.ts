import Query, { ChainableQuery, defaultClientConfiguration } from "./query";
import * as Constraints from "./constraints";
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

export {
  Query,
  ChainableQuery,
  Constraints,
  Logger,
  ConsoleTransport,
  LogLevel,
  Identity,
  // exposes a default client configuration that reverts the unexpected typeParsers
  // behaviour of the slonik library. The default typeParsers behaviour means that
  // timestamps/dates are serialized out of the db as ISO8601 strings rather than
  // native dates
  defaultClientConfiguration,
};
