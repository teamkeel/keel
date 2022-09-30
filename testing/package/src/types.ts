export type TestFunc = () => void | Promise<void>;
export type TestName = string;

const matcherTypes = <const> ['toEqual', 'notToEqual', 'toHaveError'];
type MatcherTypes = typeof matcherTypes[number];
type MatcherFunc = any
export type Matchers = Record<MatcherTypes, MatcherFunc> 

export interface Test {
  testName: TestName;
  fn: TestFunc;
}

export type AssertionFunction = (actual: Value) => boolean

export type Value = boolean | number | string | null | JsonArray | JsonMap;
interface JsonMap { [key: string]: Value; }
interface JsonArray extends Array<Value> {}

export interface RunnerOpts {
  // The port + host of the go host http server
  parentPort: number;
  host: string;

  // Shows more detailed logs about reporting of
  // test results, test pattern includes/excludes
  debug?: boolean;

  // A regex pattern to filter test case names
  pattern?: string;

  // The current filepath being processed
  filePath: string;
}

export type ScalarTypes = "string" | "boolean" | "date" | "number";

export type ModelDefinition<T> = Record<keyof T, ScalarTypes>;
