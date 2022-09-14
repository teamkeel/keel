export type TestFunc = () => void | Promise<void>;
export type TestName = string;

export interface Test {
  testName: TestName;
  fn: TestFunc;
}

export interface RunnerOpts {
  // The port + host of the go host http server
  parentPort: number;
  host: string;

  // Shows more detailed logs about reporting of
  // test results, test pattern includes/excludes
  debug?: boolean;

  // A regex pattern to filter test case names
  pattern?: string;
}

export type ScalarTypes = "string" | "boolean" | "date" | "number";

export type ModelDefinition<T> = Record<keyof T, ScalarTypes>;
