import { Config, CustomFunction, Payload, Functions } from "../types";

// Indicates a custom function did not return any value
class NoResultError extends Error {}

class NotFoundError extends Error {}

// Generic handler function that is agnostic to runtime environment (http or lambda)
// to execute a custom function correctly based on a path and a request payload
// If an error occurs during execution of the function, then an error is thrown, and
// should be handled accordingly in the caller.
const handler = async (path: string, payload: Payload, config: Config) => {
  const { api, functions } = config;

  const fn = matchPathToFunction(path, functions);
  const result = await fn.call(payload, api);

  if (!result) {
    // no result returned from custom function
    throw new NoResultError("no result returned from custom function");
  }

  return result;
};

const matchPathToFunction = (
  path: string,
  functions: Functions
): CustomFunction => {
  const normalisedPath = path.replace(/\//, "");

  if (!(normalisedPath in functions)) {
    throw new NotFoundError(
      `no matching function found for path ${path}`
    );
  }

  return functions[normalisedPath];
};

export default handler;
